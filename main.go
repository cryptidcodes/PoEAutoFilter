package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Example custom override block (edit as needed)
const customOverrideBlock string = `
#= Custom Override Block Start =
Hide
BaseType "Muttering" "Whispering" "Weeping" "Wailing" "Divine Vessel"
#= Custom Override Block End =
`

var CurrentLeague string = "Mercenaries" // Default league

var ChaosPrice float64 = 1.0   // Initial value for Chaos price
var ExaltedPrice float64 = 0.0 // Initial value for Exalted price
var DivinePrice float64 = 0.0  // Initial value for Divine price

var styleMap = map[string][]string{
	"Sub 1 Chaos": {
		"SetFontSize 45\n",
		"SetTextColor 0 0 0 255\n",
		"SetBorderColor 0 0 0 255\n",
		"SetBackgroundColor 213 159 0 255\n",
	},
	"1 Chaos": {
		"SetFontSize 45\n",
		"SetTextColor 0 0 0 255\n",
		"SetBorderColor 0 0 0 255\n",
		"SetBackgroundColor 249 150 25 255\n",
		"PlayAlertSound 2 300\n",
		"PlayEffect White\n",
		"MinimapIcon 2 White Circle\n",
	},
	"5 Chaos": {
		"SetFontSize 45\n",
		"SetTextColor 0 0 0 255\n",
		"SetBorderColor 0 0 0 255\n",
		"SetBackgroundColor 240 90 35 255\n",
		"PlayAlertSound 2 300\n",
		"PlayEffect Yellow\n",
		"MinimapIcon 1 Yellow Circle\n",
	},
	"1 Exalted": {
		"SetFontSize 45\n",
		"SetTextColor 255 255 255 255\n",
		"SetBorderColor 255 255 255 255\n",
		"SetBackgroundColor 240 90 35 255\n",
		"PlayAlertSound 1 300\n",
		"PlayEffect Red\n",
		"MinimapIcon 0 Red Circle\n",
	},
	"1 Divine": {
		"SetFontSize 45\n",
		"SetTextColor 255 0 0 255\n",
		"SetBorderColor 255 0 0 255\n",
		"SetBackgroundColor 255 255 255 255\n",
		"PlayAlertSound 6 300\n",
		"PlayEffect Red\n",
		"MinimapIcon 0 Red Star\n",
	},
}

var typeSlice = []string{"Currency", "Fragments", "Scarabs", "Fossils", "Essences"}

// This is the entry point for the Path of Exile Auto Filter application.
type Currency struct {
	CurrencyTypeName string `json:"currencyTypeName"`
	Pay              struct {
		ID                int       `json:"id"`
		LeagueID          int       `json:"league_id"`
		PayCurrencyID     int       `json:"pay_currency_id"`
		GetCurrencyID     int       `json:"get_currency_id"`
		SampleTimeUtc     time.Time `json:"sample_time_utc"`
		Count             int       `json:"count"`
		Value             float64   `json:"value"`
		DataPointCount    int       `json:"data_point_count"`
		IncludesSecondary bool      `json:"includes_secondary"`
		ListingCount      int       `json:"listing_count"`
	} `json:"pay"`
	Receive struct {
		ID                int       `json:"id"`
		LeagueID          int       `json:"league_id"`
		PayCurrencyID     int       `json:"pay_currency_id"`
		GetCurrencyID     int       `json:"get_currency_id"`
		SampleTimeUtc     time.Time `json:"sample_time_utc"`
		Count             int       `json:"count"`
		Value             float64   `json:"value"`
		DataPointCount    int       `json:"data_point_count"`
		IncludesSecondary bool      `json:"includes_secondary"`
		ListingCount      int       `json:"listing_count"`
	} `json:"receive"`
	PaySparkLine struct {
		Data        []interface{} `json:"data"`
		TotalChange float64       `json:"totalChange"`
	} `json:"paySparkLine"`
	ReceiveSparkLine struct {
		Data        []float64 `json:"data"`
		TotalChange float64   `json:"totalChange"`
	} `json:"receiveSparkLine"`
	ChaosEquivalent           float64 `json:"chaosEquivalent"`
	LowConfidencePaySparkLine struct {
		Data        []interface{} `json:"data"`
		TotalChange float64       `json:"totalChange"`
	} `json:"lowConfidencePaySparkLine"`
	LowConfidenceReceiveSparkLine struct {
		Data        []float64 `json:"data"`
		TotalChange float64   `json:"totalChange"`
	} `json:"lowConfidenceReceiveSparkLine"`
	DetailsID string `json:"detailsId"`
}

type Item struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	BaseType  string `json:"baseType"`
	StackSize int    `json:"stackSize"`
	ItemClass int    `json:"itemClass"`
	Sparkline struct {
		Data        []float64 `json:"data"`
		TotalChange float64   `json:"totalChange"`
	} `json:"sparkline"`
	LowConfidenceSparkline struct {
		Data        []float64 `json:"data"`
		TotalChange float64   `json:"totalChange"`
	} `json:"lowConfidenceSparkline"`
	ImplicitModifiers []interface{} `json:"implicitModifiers"`
	ExplicitModifiers []struct {
		Text     string `json:"text"`
		Optional bool   `json:"optional"`
	} `json:"explicitModifiers"`
	FlavourText  string        `json:"flavourText"`
	ChaosValue   float64       `json:"chaosValue"`
	ExaltedValue float64       `json:"exaltedValue"`
	DivineValue  float64       `json:"divineValue"`
	Count        int           `json:"count"`
	DetailsID    string        `json:"detailsId"`
	TradeInfo    []interface{} `json:"tradeInfo"`
	ListingCount int           `json:"listingCount"`
}

type Config struct {
	FilePath string
	League   string
	// Sub1cMult is the multiplier for sub 1 Chaos items threshold
	Sub1cMult float64
}

type CurrencyResponse struct {
	Lines []Currency `json:"lines"`
}

type ItemResponse struct {
	Lines []Item `json:"lines"`
}

func fetchCurrencyValues(league string, itemType string) ([]Currency, error) {
	url := fmt.Sprintf("https://poe.ninja/api/data/currencyoverview?league=%s&type=%s", league, itemType)
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data CurrencyResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data.Lines, nil
}

func fetchItemValues(league string, itemType string) ([]Item, error) {
	url := fmt.Sprintf("https://poe.ninja/api/data/itemoverview?league=%s&type=%s", league, itemType)
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data ItemResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data.Lines, nil
}

// Returns the filter blocks as a single string
func writeFilterBlocks(valueMap map[string]map[string]float64, sub1cMult float64) string {
	var buf bytes.Buffer
	for i := 0; i < len(typeSlice); i++ {
		for name, value := range valueMap[typeSlice[i]] {
			if strings.Contains(name, "Muttering") || strings.Contains(name, "Whispering") || strings.Contains(name, "Weeping") || strings.Contains(name, "Wailing") {
				continue
			}
			buf.WriteString("\n## Half Div Tier ##\n")
			buf.WriteString("Show\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
			stacksize := int(0.5 * DivinePrice / value)
			if 0.5*DivinePrice != value {
				stacksize += 1
			}
			buf.WriteString(fmt.Sprintf("StackSize >= %v\n", int(stacksize)))
			for _, style := range styleMap["1 Divine"] {
				buf.WriteString(style)
			}
			buf.WriteString("\n## Exalted Tier ##\n")
			buf.WriteString("Show\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
			stacksize = int(ExaltedPrice / value)
			if ExaltedPrice != value {
				stacksize += 1
			}
			buf.WriteString(fmt.Sprintf("StackSize >= %v\n", int(stacksize)))
			for _, style := range styleMap["1 Exalted"] {
				buf.WriteString(style)
			}
			buf.WriteString("\n## 5 Chaos Tier ##\n")
			buf.WriteString("Show\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
			stacksize = int(5 * ChaosPrice / value)
			if 5*ChaosPrice != value {
				stacksize += 1
			}
			buf.WriteString(fmt.Sprintf("StackSize >= %v\n", int(stacksize)))
			for _, style := range styleMap["5 Chaos"] {
				buf.WriteString(style)
			}
			buf.WriteString("\n## 1 Chaos Tier ##\n")
			buf.WriteString("Show\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
			stacksize = int(ChaosPrice / value)
			if ChaosPrice != value {
				stacksize += 1
			}
			buf.WriteString(fmt.Sprintf("StackSize >= %v\n", int(stacksize)))
			for _, style := range styleMap["1 Chaos"] {
				buf.WriteString(style)
			}

			// Disable sub 1 Chaos items
			buf.WriteString("\n## Sub 1 Chaos Tier ##\n")
			buf.WriteString("Show\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
			buf.WriteString(fmt.Sprintf("StackSize >= %v\n", int(sub1cMult*ChaosPrice/value)+1))
			for _, style := range styleMap["Sub 1 Chaos"] {
				buf.WriteString(style)
			}
			buf.WriteString("\nHide\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
		}
	}
	return buf.String()
}

// Helper to update the filter file
func updateFilterFile(filename string, blocks ...string) error {
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	// Find first "#=="
	idx := bytes.Index(content, []byte("#=="))
	if idx != -1 {
		content = content[idx:]
	}
	// Compose new content
	var buf bytes.Buffer
	for _, block := range blocks {
		buf.WriteString(block)
	}
	buf.Write(content)
	// Write back to file
	return os.WriteFile(filename, buf.Bytes(), 0644)
}

// Reads config.txt and returns a Config struct
func readConfig(filename string) (Config, error) {
	cfg := Config{
		FilePath:  "",
		Sub1cMult: 0.5, // default
	}
	file, err := os.Open(filename)
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "FilePath= ") {
			cfg.FilePath = strings.TrimPrefix(line, "FilePath= ")
		}
		if strings.HasPrefix(line, "Sub1cMult= ") {
			val := strings.TrimPrefix(line, "Sub1cMult= ")
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				cfg.Sub1cMult = f
			}
		}
		if strings.HasPrefix(line, "League= ") {
			cfg.League = strings.TrimPrefix(line, "League= ")
		}
	}
	return cfg, scanner.Err()
}

func ensureConfigFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		defaultConfig := "FilePath=Path\\To\\Your\\Filter.filter\nSub1cMult=0.5\n"
		return os.WriteFile(filename, []byte(defaultConfig), 0644)
	}
	return nil
}

func main() {
	const configFile = "config.txt"
	for {
		fmt.Println("Hello, Path of Exile Auto Filter!")

		// Ensure config.txt exists
		if err := ensureConfigFile(configFile); err != nil {
			fmt.Println("Error creating config.txt:", err)
			time.Sleep(time.Minute)
			continue
		}

		// Read config at the start of each loop
		cfg, err := readConfig(configFile)
		if err != nil {
			fmt.Println("Error reading config.txt:", err)
			time.Sleep(time.Minute)
			continue
		}

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
		fmt.Printf("Fetching item values for fragments type: Fragment\n")
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
		fmt.Printf("Fetching item values for scarabs type: Scarab\n")
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
		fmt.Printf("Fetching item values for fossils type: Fossil\n")
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
		fmt.Printf("Fetching item values for essences type: Essence\n")
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

		filter := writeFilterBlocks(valueMap, sub1cMult)
		err = updateFilterFile(filePath, customOverrideBlock, filter)
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
