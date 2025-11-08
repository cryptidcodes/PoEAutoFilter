package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	FilePath string
	League   string
	// Sub1cMult is the multiplier for sub 1 Chaos items threshold
	Sub1cMult float64
	Override  string
	Styles    map[string][]string
}

// ParseConfig reads and parses the given config file into a Config struct.
func ParseConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	cfg := Config{
		Sub1cMult: 1, // default
		Styles:    make(map[string][]string),
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
		case "League":
			if line != "" {
				cfg.League = line
			}
		case "Sub1cMult":
			if line != "" {
				f, err := strconv.ParseFloat(line, 64)
				if err == nil {
					cfg.Sub1cMult = f
				}
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
