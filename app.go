package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AppState holds the runtime state of the application
type AppState struct {
	ChaosPrice   float64
	ExaltedPrice float64
	DivinePrice  float64
	IsRunning    bool
	mu           sync.RWMutex
}

// App holds the main application dependencies and state
type App struct {
	Config     Config
	ConfigPath string
	LogFunc    func(string)
	State      *AppState

	// Context for managing the bot lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	// Mutex to protect bot lifecycle
	botMu sync.Mutex
}

// NewApp creates a new application instance
func NewApp(configPath string, logFunc func(string)) (*App, error) {
	// Initialize with default config
	cfg := Config{League: "Standard"}

	// Try to load config
	loadedCfg, err := LoadConfig(configPath)
	if err == nil {
		cfg = loadedCfg
	} else {
		if logFunc != nil {
			logFunc(fmt.Sprintf("Warning: Could not load config: %v. Using defaults.\n", err))
		}
	}

	return &App{
		Config:     cfg,
		ConfigPath: configPath,
		LogFunc:    logFunc,
		State: &AppState{
			ChaosPrice: 1.0,
			IsRunning:  false,
		},
	}, nil
}

// UpdateConfig updates the app's configuration safely
func (a *App) UpdateConfig(newCfg Config) {
	a.botMu.Lock()
	defer a.botMu.Unlock()
	a.Config = newCfg
	// Should also save to disk
	if err := SaveConfig(newCfg, a.ConfigPath); err != nil {
		a.Log(fmt.Sprintf("Error saving config: %v\n", err))
	} else {
		a.Log("Configuration saved.\n")
	}
}

// Log is a helper to write to the logger
func (a *App) Log(msg string) {
	if a.LogFunc != nil {
		a.LogFunc(msg)
	}
}

// StartBot starts the automation loop
func (a *App) StartBot() {
	a.botMu.Lock()
	defer a.botMu.Unlock()

	if a.State.IsRunning {
		a.Log("Bot is already running.\n")
		return
	}

	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.State.mu.Lock()
	a.State.IsRunning = true
	a.State.mu.Unlock()

	a.Log("Starting AutoFilter Bot...\n")
	go a.runBotLoop(a.ctx)
}

// StopBot stops the automation loop
func (a *App) StopBot() {
	a.botMu.Lock()
	defer a.botMu.Unlock()

	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}

	a.State.mu.Lock()
	a.State.IsRunning = false
	a.State.mu.Unlock()

	a.Log("Bot stopped.\n")
}

// runBotLoop is the main loop, replacing the old runBot
func (a *App) runBotLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Default wait time
	defer ticker.Stop()

	// Run immediately once
	a.processFilterUpdate(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.processFilterUpdate(ctx)
		}
	}
}

// processFilterUpdate performs one iteration of fetching prices and updating the filter
func (a *App) processFilterUpdate(ctx context.Context) {
	// Use a local copy of config to avoid races if it changes mid-update
	// Realistically we might want a Read Lock here if Config can change
	cfg := a.Config

	a.Log("Path of Exile Auto Filter Check...\n")
	a.Log("League: " + cfg.League + "\n")

	// 1. Fetch Currency
	a.Log("Fetching Currency prices...\n")
	currencyItems, err := fetchCurrencyValues(cfg.League, "Currency") // Update this function to take context?
	if err != nil {
		a.Log(fmt.Sprintf("Error fetching currency: %v\n", err))
		return
	}

	// Process Currency to update global prices (Chaos, Exalt, Divine)
	currencyMap := make(map[string]float64)
	for _, item := range currencyItems {
		currencyMap[item.CurrencyTypeName] = item.ChaosEquivalent
	}

	a.State.mu.Lock()
	a.State.ExaltedPrice = currencyMap["Exalted Orb"]
	a.State.DivinePrice = currencyMap["Divine Orb"]
	// Chaos is usually 1, but we can store it if needed
	a.State.ChaosPrice = 1.0

	exPrice := a.State.ExaltedPrice
	divPrice := a.State.DivinePrice
	a.State.mu.Unlock()

	a.Log(fmt.Sprintf("Current Prices: Divine: %.1fc, Exalt: %.1fc\n", divPrice, exPrice))

	// 2. Fetch other items (Fragments, Scarabs, etc.) - Could be parallelized
	// For now, keep sequential but organized

	valueMap := make(map[string]map[string]float64)
	valueMap["Currency"] = currencyMap

	typesToFetch := []struct {
		Name     string
		APIType  string // "itemoverview" or "currencyoverview" logic
		Category string
	}{
		{"Fragments", "currency", "Fragment"},
		{"Scarabs", "item", "Scarab"},
		{"Fossils", "item", "Fossil"},
		{"Essences", "item", "Essence"},
	}

	for _, t := range typesToFetch {
		// Check context before each heavy network op
		if ctx.Err() != nil {
			return
		}

		a.Log(fmt.Sprintf("Fetching %s...\n", t.Name))
		var priceMap map[string]float64
		var err error

		if t.Category == "Fragment" {
			// Fragments endpoint is currencyoverview
			items, e := fetchCurrencyValues(cfg.League, t.Category)
			if e == nil {
				priceMap = make(map[string]float64)
				for _, i := range items {
					priceMap[i.CurrencyTypeName] = i.ChaosEquivalent
				}
			}
			err = e
		} else {
			// Others are itemoverview
			items, e := fetchItemValues(cfg.League, t.Category)
			if e == nil {
				priceMap = make(map[string]float64)
				for _, i := range items {
					priceMap[i.Name] = i.ChaosValue
				}
			}
			err = e
		}

		if err != nil {
			a.Log(fmt.Sprintf("Error fetching %s: %v\n", t.Name, err))
			continue
		}
		valueMap[t.Name] = priceMap
	}

	// 3. Generate Filter
	a.Log("Generating filter...\n")
	// We need to pass the prices manually since we removed globals
	prices := PriceTable{
		Exalted: exPrice,
		Divine:  divPrice,
	}

	filterContent := writeFilterBlocks(cfg, valueMap, prices)

	err = updateFilterFile(cfg.BaseFilePath, cfg.FilePath, cfg.Override, filterContent)
	if err != nil {
		a.Log(fmt.Sprintf("Error updating filter file: %v\n", err))
		return
	}

	a.Log(fmt.Sprintf("Filter updated successfully at %s\n", time.Now().Format(time.Kitchen)))
}

// SetupLogger configures the global logger to write to a file and stdout
func SetupLogger() (*os.File, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exPath := filepath.Dir(ex)
	logPath := filepath.Join(exPath, "debug.log")

	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("error opening log file: %v", err)
	}

	// Write to both file and stdout
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("=== Session Started: %s ===", time.Now().Format(time.RFC3339))
	return f, nil
}

// RunGUI is a variable that can be reassigned by platform-specific files.
// By default, it prints a message that GUI is not available.
var RunGUI = func(app *App) {
	fmt.Println("GUI is only available on Windows.")
	fmt.Println("Running in headless mode (if implemented) or exiting.")
}
