package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

type FilterAction struct {
	Type   string   `json:"type"`   // e.g., "SetFontSize", "SetTextColor"
	Values []string `json:"values"` // e.g., ["45"], ["255", "0", "0"]
}

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

type Style struct {
	Name    string         `json:"name"`
	Actions []FilterAction `json:"actions"`
}

func (s Style) ToFilterLines() string {
	var lines []string
	for _, action := range s.Actions {
		lines = append(lines, action.ToFilterLine())
	}
	return strings.Join(lines, "\n")
}

type Tier struct {
	Name      string  `json:"name"`
	Value     float64 `json:"value"`
	Currency  string  `json:"currency"`  // "Chaos", "Exalted", "Divine"
	StyleName string  `json:"styleName"` // Reference to Style.Name
}

type Config struct {
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

// LoadConfig attempts to load config from JSON, falling back to legacy text format.
func LoadConfig() (Config, error) {
	var cfg Config

	// Try loading from JSON first
	if _, err := os.Stat(ConfigJSON); err == nil {
		data, err := os.ReadFile(ConfigJSON)
		if err != nil {
			return Config{}, err
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, err
		}
	} else if _, err := os.Stat(ConfigText); err == nil {
		// Fallback to legacy
		cfg, err = parseLegacyConfig(ConfigText)
		if err != nil {
			return Config{}, err
		}
	} else {
		// Defaults
		cfg = Config{
			League: "Standard",
		}
	}

	// Migration logic: styles -> styleLibrary and tiers
	if (len(cfg.StyleLibrary) == 0 && len(cfg.Styles) > 0) || len(cfg.Tiers) == 0 {
		migrateStyles(&cfg)
		// Save migrated config to JSON
		SaveConfig(cfg)
	}

	return cfg, nil
}

// migrateStyles converts the old Styles map into the new StyleLibrary and Tiers.
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

// SaveConfig saves the configuration to config.json
func SaveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigJSON, data, 0644)
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
