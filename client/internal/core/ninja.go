package core

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

// NinjaItem represents an item definition from the poe.ninja API.
type NinjaItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NinjaLine represents the market data for an item from the poe.ninja API.
type NinjaLine struct {
	ID           string  `json:"id"`
	PrimaryValue float64 `json:"primaryValue"` // Value in the primary currency (e.g. divine)
}

// NinjaCore holds the core metadata from the poe.ninja API response,
// including conversion rates from the primary currency to others.
type NinjaCore struct {
	Rates     map[string]float64 `json:"rates"`
	Primary   string             `json:"primary"`
	Secondary string             `json:"secondary"`
}

// NinjaResponse is the top-level JSON structure returned by the poe.ninja API.
type NinjaResponse struct {
	Core  NinjaCore   `json:"core"`
	Lines []NinjaLine `json:"lines"`
	Items []NinjaItem `json:"items"`
}

// PriceEntry connects the human-readable item name to its converted value.
type PriceEntry struct {
	Name  string
	Value float64
}

// FetchPrices retrieves data from the new poe.ninja v2 API endpoint.
// It constructs the URL using values from .env if available.
func FetchPrices(baseURL, gameVersion, league, itemType string) (map[string]float64, error) {
	// Attempt to load .env, but don't fail if it doesn't exist
	_ = godotenv.Load()

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

	gameVersionStr := "poe1"
	if gameVersion == "poe2" {
		gameVersionStr = "poe2"
	}

	// Construct the dynamic path. Note: poe.ninja API expects /poe1/ or /poe2/ before /api/economy
	// If the server proxy proxies exactly what it's given, we must send /<gameVersion>/api/...
	path := fmt.Sprintf("/%s%s%s%s?league=%s&type=%s", gameVersionStr, apiURL, priceSource, currentPrice, url.QueryEscape(league), url.QueryEscape(itemType))
	requestURL := baseURL + path

	log.Printf("[ninja] Fetching prices: %s", requestURL)

	resp, err := http.Get(requestURL)
	if err != nil {
		log.Printf("[ninja] HTTP error fetching prices: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ninja] Non-200 status code: %d for %s", resp.StatusCode, requestURL)
		return nil, fmt.Errorf("poe.ninja returned status %d for %s", resp.StatusCode, requestURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ninja] Error reading response body: %v", err)
		return nil, err
	}

	log.Printf("[ninja] Raw response body for %s/%s:\n%s", league, itemType, string(body))

	var data NinjaResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("[ninja] Error parsing JSON: %v", err)
		return nil, err
	}

	// Map ID -> Name
	nameMap := make(map[string]string)
	for _, item := range data.Items {
		nameMap[item.ID] = item.Name
	}

	// Determine target currency and conversion rate based on game version.
	// primaryValue is denominated in core.primary (typically "divine").
	// poe1: convert to chaos equivalent
	// poe2: convert to exalted equivalent
	targetCurrency := "chaos"
	if gameVersion == "poe2" {
		targetCurrency = "exalted"
	}

	conversionRate := 1.0
	if data.Core.Rates != nil {
		if rate, ok := data.Core.Rates[targetCurrency]; ok {
			conversionRate = rate
			log.Printf("[ninja] Converting from %s to %s using rate %.4f", data.Core.Primary, targetCurrency, rate)
		} else {
			log.Printf("[ninja] Warning: no conversion rate found for %s in core.rates, using primaryValue as-is", targetCurrency)
		}
	} else {
		log.Printf("[ninja] Warning: core.rates is empty, using primaryValue as-is")
	}

	// Map Name -> converted value
	priceMap := make(map[string]float64)
	for _, line := range data.Lines {
		if name, ok := nameMap[line.ID]; ok {
			convertedValue := line.PrimaryValue * conversionRate
			priceMap[name] = convertedValue
			log.Printf("[ninja] Price: %s = %.4f %s (primaryValue=%.6f × rate=%.4f)", name, convertedValue, targetCurrency, line.PrimaryValue, conversionRate)
		} else {
			log.Printf("[ninja] Warning: no item name found for line ID %s (primaryValue=%.6f)", line.ID, line.PrimaryValue)
		}
	}

	log.Printf("[ninja] Received %d items for %s/%s (converted to %s)", len(priceMap), league, itemType, targetCurrency)
	return priceMap, nil
}
