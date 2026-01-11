package main

import (
	"fmt"
	"time"
)

var CurrentLeague string = "Standard" // Default league
var ChaosPrice float64 = 1.0          // Initial value for Chaos price
var ExaltedPrice float64 = 0.0        // Initial value for Exalted price
var DivinePrice float64 = 0.0         // Initial value for Divine price

var typeSlice = []string{"Currency", "Fragments", "Scarabs", "Fossils", "Essences"}

// main is the entry point. It now launches the GUI.
func main() {
	ShowGUI()
}

// runBot contains the main automation logic, extracted from the old main function.
// It runs indefinitely.
func runBot(cfg Config, logger func(string)) {
	for {
		logger("Path of Exile Auto Filter\n")
		// In CLI version we re-read config here. In GUI version, config is locked in.
		// If dynamic updates are needed, we would re-load 'cfg' here.

		logger("Configured successfully with League: " + cfg.League + "\n")

		filePath := cfg.FilePath

		// Fetch item values for the specified league and currency item type
		logger("Fetching item values for currency type: Currency\n")
		items, err := fetchCurrencyValues(cfg.League, "Currency")
		if err != nil {
			logger(fmt.Sprintf("Error fetching items: %v\n", err))
			logger("Retrying in 1 minute...\n")
			time.Sleep(time.Minute)
			continue
		}
		if len(items) == 0 {
			logger("No items found.\n")
			time.Sleep(time.Minute)
			continue
		}
		logger(fmt.Sprintf("Found %d items\n", len(items)))
		currencyValues := make(map[string]float64)
		for i := 0; i < len(items); i++ {
			currencyValues[items[i].CurrencyTypeName] = items[i].ChaosEquivalent
		}

		ExaltedPrice = currencyValues["Exalted Orb"]
		DivinePrice = currencyValues["Divine Orb"]

		logger("Current Prices:\n")
		logger(fmt.Sprintf("Chaos Orb: %fc\n", ChaosPrice))
		logger(fmt.Sprintf("Exalted Orb: %fc\n", ExaltedPrice))
		logger(fmt.Sprintf("Divine Orb: %fc\n", DivinePrice))

		// Fragments
		logger("Fetching item values for type: Fragment\n")
		fragments, err := fetchCurrencyValues(cfg.League, "Fragment")
		if err != nil {
			logger(fmt.Sprintf("Error fetching fragments: %v\n", err))
			continue
		}
		if len(fragments) == 0 {
			logger("No fragments found.\n")
			continue
		}
		logger(fmt.Sprintf("Found %d fragments\n", len(fragments)))
		fragmentValues := make(map[string]float64)
		for i := 0; i < len(fragments); i++ {
			fragmentValues[fragments[i].CurrencyTypeName] = fragments[i].ChaosEquivalent
		}

		// Scarabs
		logger("Fetching item values for type: Scarab\n")
		scarabs, err := fetchItemValues(cfg.League, "Scarab")
		if err != nil {
			logger(fmt.Sprintf("Error fetching scarabs: %v\n", err))
			continue
		}
		if len(scarabs) == 0 {
			logger("No scarabs found.\n")
			continue
		}
		logger(fmt.Sprintf("Found %d scarabs\n", len(scarabs)))
		scarabValues := make(map[string]float64)
		for i := 0; i < len(scarabs); i++ {
			scarabValues[scarabs[i].Name] = scarabs[i].ChaosValue
		}

		// Fossils
		logger("Fetching item values for type: Fossil\n")
		fossils, err := fetchItemValues(cfg.League, "Fossil")
		if err != nil {
			logger(fmt.Sprintf("Error fetching fossils: %v\n", err))
			continue
		}
		if len(fossils) == 0 {
			logger("No fossils found.\n")
			continue
		}
		logger(fmt.Sprintf("Found %d fossils\n", len(fossils)))
		fossilValues := make(map[string]float64)
		for i := 0; i < len(fossils); i++ {
			fossilValues[fossils[i].Name] = fossils[i].ChaosValue
		}

		// Essences
		logger("Fetching item values for type: Essence\n")
		essences, err := fetchItemValues(cfg.League, "Essence")
		if err != nil {
			logger(fmt.Sprintf("Error fetching essences: %v\n", err))
			continue
		}
		if len(essences) == 0 {
			logger("No essences found.\n")
			continue
		}
		logger(fmt.Sprintf("Found %d essences\n", len(essences)))
		essenceValues := make(map[string]float64)
		for i := 0; i < len(essences); i++ {
			essenceValues[essences[i].Name] = essences[i].ChaosValue
		}

		// Create a ValueMap to hold all the values
		valueMap := make(map[string]map[string]float64)
		valueMap["Currency"] = currencyValues
		valueMap["Fragments"] = fragmentValues
		valueMap["Scarabs"] = scarabValues
		valueMap["Fossils"] = fossilValues
		valueMap["Essences"] = essenceValues

		filter := writeFilterBlocks(cfg, valueMap)
		// OPTIONAL: Add customOverrideBlock as arg before the filter arg if you want to include it
		// TODO: Transition customOverrideBlock to a config file or GUI setting
		err = updateFilterFile(cfg.BaseFilePath, filePath, cfg.Override, filter)
		if err != nil {
			logger(fmt.Sprintf("Error updating filter file: %v\n", err))
			continue
		}
		logger("Filter file updated successfully!\n")
		logger(fmt.Sprintf("Filter blocks written to file: %s\nAt Time: %s\n", filePath, time.Now()))

		logger("Waiting 1 hour before next update...\n")
		time.Sleep(time.Hour)
	}
}
