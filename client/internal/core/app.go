package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// AppState holds the runtime state of the application including current prices
// and whether the bot loop is active. Thread-safe via RWMutex.
type AppState struct {
	ChaosPrice   float64
	ExaltedPrice float64
	DivinePrice  float64
	IsRunning    bool
	Mu           sync.RWMutex
}

// App holds the main application dependencies and state.
// It is the central coordinator between config, price fetching, and filter generation.
// Both the GUI and CLI entry points create and operate on an App instance.
type App struct {
	Config     Config
	ConfigPath string
	LogFunc    func(string)
	State      *AppState

	// BaseURL for poe.ninja API requests. Defaults to DefaultBaseURL.
	// Can be overridden with --server-url flag for custom edge server deployments.
	BaseURL string

	// Context for managing the bot lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	// Mutex to protect bot lifecycle
	botMu sync.Mutex
}

// NewApp creates a new application instance with config loaded from the specified path.
// If config loading fails, defaults are used. The logFunc parameter is optional — pass nil
// for headless/CLI mode and it will be set later or left as a no-op.
func NewApp(configPath string, logFunc func(string)) (*App, error) {
	log.Printf("[app] Initializing app with config: %s", configPath)

	// Initialize with default config
	cfg := Config{League: "Standard"}

	// Try to load config
	loadedCfg, err := LoadConfig(configPath)
	if err == nil {
		cfg = loadedCfg
	} else {
		log.Printf("[app] Warning: Could not load config: %v. Using defaults.", err)
		if logFunc != nil {
			logFunc(fmt.Sprintf("Warning: Could not load config: %v. Using defaults.\n", err))
		}
	}

	return &App{
		Config:     cfg,
		ConfigPath: configPath,
		LogFunc:    logFunc,
		BaseURL:    "https://poe.ninja",
		State: &AppState{
			ChaosPrice: 1.0,
			IsRunning:  false,
		},
	}, nil
}

// UpdateConfig updates the app's configuration safely and persists it to disk.
func (a *App) UpdateConfig(newCfg Config) {
	a.botMu.Lock()
	defer a.botMu.Unlock()
	a.Config = newCfg
	// Should also save to disk
	if err := SaveConfig(&a.Config, a.ConfigPath); err != nil {
		a.Log(fmt.Sprintf("Error saving config: %v\n", err))
	} else {
		a.Log("Configuration saved.\n")
	}
}

// Log is a helper to write to the configured logger function.
// Safe to call even if LogFunc is nil (no-op).
func (a *App) Log(msg string) {
	if a.LogFunc != nil {
		a.LogFunc(msg)
	}
}

// StartBot starts the automation loop that periodically fetches prices and updates the filter.
// If already running, this is a no-op.
func (a *App) StartBot() {
	a.botMu.Lock()
	defer a.botMu.Unlock()

	if a.State.IsRunning {
		a.Log("Bot is already running.\n")
		return
	}

	if a.BaseURL == "" {
		a.Log("Error: Server URL not configured. Set AUTOFILTER_SERVER_URL environment variable.\n")
		return
	}

	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.State.Mu.Lock()
	a.State.IsRunning = true
	a.State.Mu.Unlock()

	a.Log("Starting AutoFilter Bot...\n")
	go a.runBotLoop(a.ctx)
}

// StopBot stops the automation loop. Safe to call even if not running.
func (a *App) StopBot() {
	a.botMu.Lock()
	defer a.botMu.Unlock()

	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}

	a.State.Mu.Lock()
	a.State.IsRunning = false
	a.State.Mu.Unlock()

	a.Log("Bot stopped.\n")
}

// runBotLoop is the main loop that runs in a goroutine, performing periodic filter updates.
// It runs once immediately, then on a 1-hour ticker until the context is cancelled.
func (a *App) runBotLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Default wait time
	defer ticker.Stop()

	// Run immediately once
	a.ProcessFilterUpdate(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.ProcessFilterUpdate(ctx)
		}
	}
}

// ProcessFilterUpdate performs one iteration of fetching prices and updating the filter.
// This is exported so the CLI 'run' command can call it directly for one-shot mode.
func (a *App) ProcessFilterUpdate(ctx context.Context) {
	// Use a local copy of config to avoid races if it changes mid-update
	cfg := a.Config
	gameVersion := cfg.GameVersion
	if gameVersion == "" {
		gameVersion = "poe1"
	}

	a.Log("Path of Exile Auto Filter Check...\n")
	a.Log("League: " + cfg.League + "\n")
	a.Log("Game Version: " + gameVersion + "\n")

	// 1. Fetch Currency
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "/api/economy"
	}
	priceSource := os.Getenv("PRICE_SOURCE_EXCHANGE")
	if priceSource == "" {
		priceSource = "/exchange"
	}
	currentPrice := os.Getenv("CURRENT_PRICE")
	if currentPrice == "" {
		currentPrice = "/current/overview"
	}

	a.Log("Fetching Currency prices...\n")
	currencyMap, err := FetchPrices(a.BaseURL, gameVersion, cfg.League, "Currency")
	if err != nil {
		a.Log(fmt.Sprintf("Error fetching currency: %v\n", err))
		return
	}
	if len(currencyMap) == 0 {
		a.Log(fmt.Sprintf("Warning: No price data returned for league %q. Please verify spelling and capitalization (e.g. \"Runes of Aldur\" vs \"Runes Of Aldur\").\n", cfg.League))
	}

	// For PoE2, poe.ninja prices are in Divine Orbs.
	// Normalize them to Chaos equivalents by dividing by the Chaos Orb price.
	chaosPriceInEx := 1.0
	if gameVersion == "poe2" {
		if p, ok := currencyMap["Chaos Orb"]; ok && p > 0 {
			chaosPriceInEx = p
		} else {
			chaosPriceInEx = 0.1 // fallback
		}
		a.Log(fmt.Sprintf("PoE2 Base Currency Normalization: 1 Chaos = %.4f Exalted\n", chaosPriceInEx))

		for k, v := range currencyMap {
			currencyMap[k] = v / chaosPriceInEx
		}
	}

	a.State.Mu.Lock()
	a.State.ExaltedPrice = currencyMap["Exalted Orb"]
	a.State.DivinePrice = currencyMap["Divine Orb"]
	a.State.ChaosPrice = 1.0

	exPrice := a.State.ExaltedPrice
	divPrice := a.State.DivinePrice
	a.State.Mu.Unlock()

	a.Log(fmt.Sprintf("Current Prices: Divine: %.1fc, Exalt: %.3fc\n", divPrice, exPrice))

	// 2. Fetch other stackable currency (Fragments, Scarabs, etc.)
	valueMap := make(map[string]map[string]float64)
	valueMap["Currency"] = currencyMap

	var typesToFetch []struct {
		Name     string
		Category string
	}

	switch gameVersion {
	case "poe2":
		typesToFetch = []struct {
			Name     string
			Category string
		}{
			{"Fragments", "Fragments"},
			{"Essences", "Essences"},
			{"Breach", "Breach"},
		}
	default:
		typesToFetch = []struct {
			Name     string
			Category string
		}{
			{"Fragments", "Fragment"},
			{"Scarabs", "Scarab"},
			{"Fossils", "Fossil"},
			{"Resonators", "Resonator"},
			{"Essences", "Essence"},
		}
	}

	for _, t := range typesToFetch {
		// Check context before each heavy network op
		if ctx.Err() != nil {
			return
		}

		a.Log(fmt.Sprintf("Fetching %s...\n", t.Name))

		priceMap, err := FetchPrices(a.BaseURL, gameVersion, cfg.League, t.Category)
		if err != nil {
			a.Log(fmt.Sprintf("Error fetching %s: %v\n", t.Name, err))
			continue
		}
		if len(priceMap) == 0 {
			a.Log(fmt.Sprintf("Warning: No data returned for %s under league %q.\n", t.Name, cfg.League))
		}
		if gameVersion == "poe2" {
			for k, v := range priceMap {
				priceMap[k] = v / chaosPriceInEx
			}
		}
		valueMap[t.Name] = priceMap
	}

	// 3. Generate Filter
	a.Log("Generating filter...\n")
	prices := PriceTable{
		Exalted: exPrice,
		Divine:  divPrice,
	}

	filterContent := WriteFilterBlocks(cfg, valueMap, prices)

	err = UpdateFilterFile(cfg.BaseFilePath, cfg.FilePath, cfg.Override, filterContent)
	if err != nil {
		a.Log(fmt.Sprintf("Error updating filter file: %v\n", err))
		return
	}

	a.Log(fmt.Sprintf("Filter updated successfully at %s\n", time.Now().Format(time.Kitchen)))
}
