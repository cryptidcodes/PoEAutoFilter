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

// This is the entry point for the Path of Exile Auto Filter application.

func main() {
	for {
		fmt.Println("Path of Exile Auto Filter")
		// Read config at the start of each loop
		fmt.Println("Reading config...")
		cfg, err := ParseConfig("config.txt")
		if err != nil {
			fmt.Println("Error:", err)
			time.Sleep(time.Second * 10)
			return
		}
		fmt.Println("Configured successfully")

		filePath := cfg.FilePath
		sub1cMult := cfg.Sub1cMult

		// Fetch item values for the specified league and currency item type
		fmt.Printf("Fetching item values for currency type: Currency\n")
		items, err := fetchCurrencyValues(cfg.League, "Currency")
		if err != nil {
			fmt.Println("Error fetching items:", err)
			continue
		}
		if len(items) == 0 {
			fmt.Println("No items found.")
			continue
		}
		fmt.Printf("Found %d items\n", len(items))
		currencyValues := make(map[string]float64)
		for i := 0; i < len(items); i++ {
			currencyValues[items[i].CurrencyTypeName] = items[i].ChaosEquivalent
		}

		ExaltedPrice = currencyValues["Exalted Orb"]
		DivinePrice = currencyValues["Divine Orb"]

		fmt.Printf("Current Prices:\n")
		fmt.Printf("Chaos Orb: %fc\n", ChaosPrice)
		fmt.Printf("Exalted Orb: %fc\n", ExaltedPrice)
		fmt.Printf("Divine Orb: %fc\n", DivinePrice)

		// Fragments
		fmt.Printf("Fetching item values for type: Fragment\n")
		fragments, err := fetchCurrencyValues(cfg.League, "Fragment")
		if err != nil {
			fmt.Println("Error fetching fragments:", err)
			continue
		}
		if len(fragments) == 0 {
			fmt.Println("No fragments found.")
			continue
		}
		fmt.Printf("Found %d fragments\n", len(fragments))
		fragmentValues := make(map[string]float64)
		for i := 0; i < len(fragments); i++ {
			fragmentValues[fragments[i].CurrencyTypeName] = fragments[i].ChaosEquivalent
		}

		// Scarabs
		fmt.Printf("Fetching item values for type: Scarab\n")
		scarabs, err := fetchItemValues(cfg.League, "Scarab")
		if err != nil {
			fmt.Println("Error fetching scarabs:", err)
			continue
		}
		if len(scarabs) == 0 {
			fmt.Println("No scarabs found.")
			continue
		}
		fmt.Printf("Found %d scarabs\n", len(scarabs))
		scarabValues := make(map[string]float64)
		for i := 0; i < len(scarabs); i++ {
			scarabValues[scarabs[i].Name] = scarabs[i].ChaosValue
		}

		// Fossils
		fmt.Printf("Fetching item values for type: Fossil\n")
		fossils, err := fetchItemValues(cfg.League, "Fossil")
		if err != nil {
			fmt.Println("Error fetching fossils:", err)
			continue
		}
		if len(fossils) == 0 {
			fmt.Println("No fossils found.")
			continue
		}
		fmt.Printf("Found %d fossils\n", len(fossils))
		fossilValues := make(map[string]float64)
		for i := 0; i < len(fossils); i++ {
			fossilValues[fossils[i].Name] = fossils[i].ChaosValue
		}

		// Essences
		fmt.Printf("Fetching item values for type: Essence\n")
		essences, err := fetchItemValues(cfg.League, "Essence")
		if err != nil {
			fmt.Println("Error fetching essences:", err)
			continue
		}
		if len(essences) == 0 {
			fmt.Println("No essences found.")
			continue
		}
		fmt.Printf("Found %d essences\n", len(essences))
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

		filter := writeFilterBlocks(cfg, valueMap, sub1cMult)
		// OPTIONAL: Add customOverrideBlock as arg before the filter arg if you want to include it
		// TODO: Transition customOverrideBlock to a config file or GUI setting
		err = updateFilterFile(filePath, cfg.Override, filter)
		if err != nil {
			fmt.Println("Error updating filter file:", err)
			continue
		}
		fmt.Println("Filter file updated successfully!")
		fmt.Printf("Filter blocks written to file: %s\nAt Time: %s\n", filePath, time.Now())

		fmt.Println("Waiting 1 hour before next update...")
		time.Sleep(time.Hour)
	}
}
