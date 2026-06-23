// Package core contains the platform-agnostic business logic for PoEAutoFilter.
// This includes configuration management, filter generation, and poe.ninja API integration.
// No platform-specific GUI imports are allowed in this package.
package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// FilterAction represents a single action line in a PoE filter style.
// For example: SetFontSize 45, SetTextColor 255 0 0 255, PlayAlertSound 6 300.
// See https://www.pathofexile.com/item-filter/about for valid action types.
type FilterAction struct {
	Type   string   `json:"type"`   // e.g., "SetFontSize", "SetTextColor"
	Values []string `json:"values"` // e.g., ["45"], ["255", "0", "0"]
}

// ToFilterLine converts a FilterAction to its PoE filter file syntax representation.
// MinimapIcon sizes are translated from descriptive strings ("Large"→"0", "Medium"→"1", "Small"→"2").
func (a FilterAction) ToFilterLine() string {
	line := a.Type
	for i, val := range a.Values {
		// Translation for descriptive values to numeric ones for PoE
		if a.Type == "MinimapIcon" && i == 0 {
			switch val {
			case "Large":
				val = "0"
			case "Medium":
				val = "1"
			case "Small":
				val = "2"
			}
		}
		line += " " + val
	}
	return line
}

// Style is a named collection of FilterActions that defines how matching items
// appear in the game (font size, colors, sounds, minimap icons, etc.).
type Style struct {
	Name    string         `json:"name"`
	Actions []FilterAction `json:"actions"`
}

// ToFilterLines converts all actions in a Style to their filter file representation,
// joined by newlines. Used for preview display and actual filter generation.
func (s Style) ToFilterLines() string {
	var lines []string
	for _, action := range s.Actions {
		lines = append(lines, action.ToFilterLine())
	}
	return strings.Join(lines, "\n")
}

// Tier defines a price threshold that maps items to a visual style.
// During filter generation, items whose price exceeds this tier's threshold
// (converted to chaos equivalent) receive the referenced style.
type Tier struct {
	Name      string  `json:"name"`
	Value     float64 `json:"value"`
	Currency  string  `json:"currency"`  // "Chaos", "Exalted", "Divine"
	StyleName string  `json:"styleName"` // Reference to Style.Name
}

// GameConfig holds sub-configuration fields unique to a game version (PoE1 or PoE2).
type GameConfig struct {
	FilePath     string  `json:"filePath"`
	BaseFilePath string  `json:"baseFilePath"`
	League       string  `json:"league"`
	Override     string  `json:"override"`
	StyleLibrary []Style `json:"styleLibrary"`
	Tiers        []Tier  `json:"tiers"`
}

// Config holds all persistent application configuration including file paths,
// league selection, style library, value tiers, and custom override rules.
type Config struct {
	GameVersion  string              `json:"gameVersion"` // "poe1" or "poe2"
	PoE1         GameConfig          `json:"poe1"`
	PoE2         GameConfig          `json:"poe2"`
	FilePath     string              `json:"filePath"`
	BaseFilePath string              `json:"baseFilePath"`
	League       string              `json:"league"`
	Override     string              `json:"override"`
	StyleLibrary []Style             `json:"styleLibrary"`
	Tiers        []Tier              `json:"tiers"`
	Styles       map[string][]string `json:"styles,omitempty"` // Legacy, for migration
}

const (
	ConfigJSON = "config.json"
	ConfigText = "config.txt"
)

// LoadConfig attempts to load config from JSON at the specified path, falling back to legacy text format.
// If neither exists, returns a default config with League="Standard".
func LoadConfig(path string) (Config, error) {
	var cfg Config

	log.Printf("[config] Loading config from path: %s", path)

	// Try loading from JSON first
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("[config] Error reading config file: %v", err)
			return Config{}, err
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Printf("[config] Error parsing JSON config: %v", err)
			return Config{}, err
		}
		log.Printf("[config] Loaded JSON config: league=%s, %d styles, %d tiers",
			cfg.League, len(cfg.StyleLibrary), len(cfg.Tiers))
	} else if _, err := os.Stat(ConfigText); err == nil {
		// Fallback to legacy
		log.Printf("[config] JSON not found, falling back to legacy config: %s", ConfigText)
		cfg, err = parseLegacyConfig(ConfigText)
		if err != nil {
			log.Printf("[config] Error parsing legacy config: %v", err)
			return Config{}, err
		}
	} else {
		// Defaults
		log.Printf("[config] No config found, using defaults")
		cfg = Config{
			League: "Standard",
		}
	}

	// Migration logic: styles -> styleLibrary and tiers
	if (len(cfg.StyleLibrary) == 0 && len(cfg.Styles) > 0) || len(cfg.Tiers) == 0 {
		log.Printf("[config] Running migration from legacy styles")
		migrateStyles(&cfg)
		// Save migrated config to JSON
		SaveConfig(&cfg, path)
	}

	if cfg.GameVersion == "" {
		cfg.GameVersion = "poe1"
	}

	// Migration to PoE1 and PoE2 structures:
	// If PoE1 has no styles and tiers, but root does, initialize PoE1 from root
	if len(cfg.PoE1.StyleLibrary) == 0 && len(cfg.PoE1.Tiers) == 0 {
		log.Printf("[config] Migrating active root config to PoE1 structure")
		cfg.PoE1 = GameConfig{
			FilePath:     cfg.FilePath,
			BaseFilePath: cfg.BaseFilePath,
			League:       cfg.League,
			Override:     cfg.Override,
			StyleLibrary: cfg.StyleLibrary,
			Tiers:        cfg.Tiers,
		}
	}

	// Ensure PoE2 has sensible defaults if empty
	if len(cfg.PoE2.StyleLibrary) == 0 && len(cfg.PoE2.Tiers) == 0 {
		log.Printf("[config] Initializing PoE2 config with default show styles")
		cfg.PoE2 = GameConfig{
			League: "Standard",
			StyleLibrary: []Style{
				{Name: "Default Show", Actions: []FilterAction{{Type: "SetFontSize", Values: []string{"32"}}}},
			},
			Tiers: []Tier{
				{Name: "1 Chaos", Value: 1.0, Currency: "Chaos", StyleName: "Default Show"},
			},
		}
	}

	// Load currently active game version into active root fields
	cfg.LoadGameConfigToActive()

	return cfg, nil
}

// migrateStyles converts the old Styles map into the new StyleLibrary and Tiers.
// This handles backward compatibility with the legacy config.txt format.
func migrateStyles(cfg *Config) {
	// Only migrate if we have old styles and no new tiers
	if len(cfg.Styles) == 0 {
		// Provide basic defaults if absolutely nothing exists
		if len(cfg.Tiers) == 0 {
			cfg.StyleLibrary = []Style{
				{Name: "Default Show", Actions: []FilterAction{{Type: "SetFontSize", Values: []string{"32"}}}},
			}
			cfg.Tiers = []Tier{
				{Name: "1 Chaos", Value: 1.0, Currency: "Chaos", StyleName: "Default Show"},
			}
		}
		return
	}

	for name, lines := range cfg.Styles {
		style := Style{Name: name}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) > 0 {
				style.Actions = append(style.Actions, FilterAction{
					Type:   parts[0],
					Values: parts[1:],
				})
			}
		}
		cfg.StyleLibrary = append(cfg.StyleLibrary, style)

		// Create a tier for it if it looks like a value tier
		tier := Tier{Name: name, StyleName: name}
		switch name {
		case "Divine":
			tier.Value = 1.0
			tier.Currency = "Divine"
		case "Exalted":
			tier.Value = 1.0
			tier.Currency = "Exalted"
		case "5 Chaos":
			tier.Value = 5.0
			tier.Currency = "Chaos"
		case "1 Chaos":
			tier.Value = 1.0
			tier.Currency = "Chaos"
		case "Sub 1 Chaos":
			tier.Value = 0.1
			tier.Currency = "Chaos"
		default:
			tier.Value = 1.0
			tier.Currency = "Chaos"
		}
		cfg.Tiers = append(cfg.Tiers, tier)
	}
	// Clear old styles after migration
	cfg.Styles = nil
}

// SaveActiveToGameConfig copies current active root configuration fields to the selected game version's sub-config.
func (c *Config) SaveActiveToGameConfig() {
	gameVersion := c.GameVersion
	if gameVersion == "" {
		gameVersion = "poe1"
	}

	gameCfg := GameConfig{
		FilePath:     c.FilePath,
		BaseFilePath: c.BaseFilePath,
		League:       c.League,
		Override:     c.Override,
		StyleLibrary: c.StyleLibrary,
		Tiers:        c.Tiers,
	}

	if gameVersion == "poe2" {
		c.PoE2 = gameCfg
	} else {
		c.PoE1 = gameCfg
	}
}

// LoadGameConfigToActive loads the configuration fields of the selected game version into the active root fields.
func (c *Config) LoadGameConfigToActive() {
	gameVersion := c.GameVersion
	if gameVersion == "" {
		gameVersion = "poe1"
	}

	var gameCfg GameConfig
	if gameVersion == "poe2" {
		gameCfg = c.PoE2
	} else {
		gameCfg = c.PoE1
	}

	c.FilePath = gameCfg.FilePath
	c.BaseFilePath = gameCfg.BaseFilePath
	c.League = gameCfg.League
	c.Override = gameCfg.Override
	c.StyleLibrary = gameCfg.StyleLibrary
	c.Tiers = gameCfg.Tiers
}

// SwitchGameVersion saves the current active configuration, switches GameVersion, and loads the target config.
func (c *Config) SwitchGameVersion(newVer string) {
	if newVer == "" {
		newVer = "poe1"
	}
	c.SaveActiveToGameConfig()
	c.GameVersion = newVer
	c.LoadGameConfigToActive()
}

// SaveConfig saves the configuration to the specified path as indented JSON.
func SaveConfig(cfg *Config, path string) error {
	log.Printf("[config] Saving config to: %s", path)
	cfg.SaveActiveToGameConfig()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("[config] Error marshaling config: %v", err)
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// --- CRUD helpers for CLI and programmatic access ---

// ListStyles returns a formatted string table of all styles in the config.
// Each style shows its name and a preview of its filter actions.
func ListStyles(cfg Config) string {
	if len(cfg.StyleLibrary) == 0 {
		return "No styles defined."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-20s  %s\n", "NAME", "ACTIONS"))
	sb.WriteString(fmt.Sprintf("%-20s  %s\n", strings.Repeat("-", 20), strings.Repeat("-", 40)))
	for _, s := range cfg.StyleLibrary {
		preview := strings.ReplaceAll(s.ToFilterLines(), "\n", "; ")
		if len(preview) > 60 {
			preview = preview[:57] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-20s  %s\n", s.Name, preview))
	}
	return sb.String()
}

// ListTiers returns a formatted string table of all tiers in the config.
func ListTiers(cfg Config) string {
	if len(cfg.Tiers) == 0 {
		return "No tiers defined."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-20s  %-10s  %-10s  %s\n", "NAME", "VALUE", "CURRENCY", "STYLE"))
	sb.WriteString(fmt.Sprintf("%-20s  %-10s  %-10s  %s\n",
		strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 10), strings.Repeat("-", 20)))
	for _, t := range cfg.Tiers {
		sb.WriteString(fmt.Sprintf("%-20s  %-10.2f  %-10s  %s\n", t.Name, t.Value, t.Currency, t.StyleName))
	}
	return sb.String()
}

// AddStyle adds a new style to the config. Returns an error if a style with the
// same name already exists.
func AddStyle(cfg *Config, style Style) error {
	for _, s := range cfg.StyleLibrary {
		if s.Name == style.Name {
			return fmt.Errorf("style %q already exists", style.Name)
		}
	}
	log.Printf("[config] Adding style: %s", style.Name)
	cfg.StyleLibrary = append(cfg.StyleLibrary, style)
	return nil
}

// RemoveStyle removes a style by name. Returns an error if not found.
func RemoveStyle(cfg *Config, name string) error {
	for i, s := range cfg.StyleLibrary {
		if s.Name == name {
			log.Printf("[config] Removing style: %s", name)
			cfg.StyleLibrary = append(cfg.StyleLibrary[:i], cfg.StyleLibrary[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("style %q not found", name)
}

// UpdateStyle replaces an existing style by name with new data. Returns an error if not found.
func UpdateStyle(cfg *Config, name string, updated Style) error {
	for i, s := range cfg.StyleLibrary {
		if s.Name == name {
			log.Printf("[config] Updating style: %s", name)
			cfg.StyleLibrary[i] = updated
			return nil
		}
	}
	return fmt.Errorf("style %q not found", name)
}

// AddTier adds a new tier to the config. Returns an error if a tier with the
// same name already exists.
func AddTier(cfg *Config, tier Tier) error {
	for _, t := range cfg.Tiers {
		if t.Name == tier.Name {
			return fmt.Errorf("tier %q already exists", tier.Name)
		}
	}
	log.Printf("[config] Adding tier: %s (%.2f %s → %s)", tier.Name, tier.Value, tier.Currency, tier.StyleName)
	cfg.Tiers = append(cfg.Tiers, tier)
	return nil
}

// RemoveTier removes a tier by name. Returns an error if not found.
func RemoveTier(cfg *Config, name string) error {
	for i, t := range cfg.Tiers {
		if t.Name == name {
			log.Printf("[config] Removing tier: %s", name)
			cfg.Tiers = append(cfg.Tiers[:i], cfg.Tiers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("tier %q not found", name)
}

// UpdateTier replaces an existing tier by name with new data. Returns an error if not found.
func UpdateTier(cfg *Config, name string, updated Tier) error {
	for i, t := range cfg.Tiers {
		if t.Name == name {
			log.Printf("[config] Updating tier: %s", name)
			cfg.Tiers[i] = updated
			return nil
		}
	}
	return fmt.Errorf("tier %q not found", name)
}

// FindStyleByName returns a pointer to a style by name, or nil if not found.
func FindStyleByName(cfg *Config, name string) *Style {
	for i, s := range cfg.StyleLibrary {
		if s.Name == name {
			return &cfg.StyleLibrary[i]
		}
	}
	return nil
}

// FindTierByName returns a pointer to a tier by name, or nil if not found.
func FindTierByName(cfg *Config, name string) *Tier {
	for i, t := range cfg.Tiers {
		if t.Name == name {
			return &cfg.Tiers[i]
		}
	}
	return nil
}

// parseLegacyConfig reads and parses the old config.txt file.
func parseLegacyConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	cfg := Config{
		Styles: make(map[string][]string),
	}

	scanner := bufio.NewScanner(file)
	var (
		currentSection string
		inStyles       bool
		builder        strings.Builder
		styleName      string
		styleLines     []string
	)

	for scanner.Scan() {
		line := scanner.Text()

		// Trim any carriage returns or trailing spaces
		line = strings.TrimRight(line, "\r ")

		// Detect top-level section headers (***Name***)
		if strings.HasPrefix(line, "***") && strings.HasSuffix(line, "***") {
			// If we were building Override section, store it
			if currentSection == "Override" {
				cfg.Override = builder.String()
				builder.Reset()
			}
			// If we were collecting a style block, store it
			if styleName != "" && len(styleLines) > 0 {
				cfg.Styles[styleName] = append([]string(nil), styleLines...)
				styleLines = nil
			}

			currentSection = strings.Trim(line, "*")
			currentSection = strings.TrimSpace(currentSection)

			// Check if entering or leaving styles section
			if currentSection == "Styles" {
				inStyles = true
			} else if inStyles && currentSection != "Styles" {
				inStyles = false
			}

			continue
		}

		// Inside Styles section
		if inStyles {
			// Style headers (###Name###)
			if strings.HasPrefix(line, "###") && strings.HasSuffix(line, "###") {
				// Save previous style before starting a new one
				if styleName != "" && len(styleLines) > 0 {
					cfg.Styles[styleName] = append([]string(nil), styleLines...)
					styleLines = nil
				}
				styleName = strings.Trim(line, "#")
				styleName = strings.TrimSpace(styleName)
				continue
			}

			// Collect style lines
			if styleName != "" && line != "" {
				styleLines = append(styleLines, line+"\n")
			}
			continue
		}

		// Handle each section content
		switch currentSection {
		case "FilePath":
			if line != "" {
				cfg.FilePath = line
			}
		case "BaseFilePath":
			if line != "" {
				cfg.BaseFilePath = line
			}
		case "League":
			if line != "" {
				cfg.League = line
			}
		case "Override":
			builder.WriteString(line + "\n")
		}
	}

	// Store any remaining collected section
	if currentSection == "Override" {
		cfg.Override = builder.String()
	}
	if styleName != "" && len(styleLines) > 0 {
		cfg.Styles[styleName] = append([]string(nil), styleLines...)
	}

	if err := scanner.Err(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
