package core

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	PrimaryValue float64 `json:"primaryValue"` // This is the chaos equivalent value
}

// NinjaResponse is the top-level JSON structure returned by the new poe.ninja API.
type NinjaResponse struct {
	Lines []NinjaLine `json:"lines"`
	Items []NinjaItem `json:"items"`
}

// PriceEntry connects the human-readable item name to its chaos value.
type PriceEntry struct {
	Name       string
	ChaosValue float64
}

// FetchPrices retrieves data from the new poe.ninja v2 API endpoint.
// It constructs the URL using values from .env if available.
func FetchPrices(baseURL, league, itemType string) (map[string]float64, error) {
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

	// Construct the dynamic path. Note: poe.ninja API expects /poe1/ before /api/economy
	// If the server proxy proxies exactly what it's given, we must send /poe1/api/...
	path := fmt.Sprintf("/poe1%s%s%s?league=%s&type=%s", apiURL, priceSource, currentPrice, league, itemType)
	url := baseURL + path

	log.Printf("[ninja] Fetching prices: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[ninja] HTTP error fetching prices: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ninja] Non-200 status code: %d for %s", resp.StatusCode, url)
		return nil, fmt.Errorf("poe.ninja returned status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ninja] Error reading response body: %v", err)
		return nil, err
	}

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

	// Map Name -> ChaosValue
	priceMap := make(map[string]float64)
	for _, line := range data.Lines {
		if name, ok := nameMap[line.ID]; ok {
			priceMap[name] = line.PrimaryValue
		}
	}

	log.Printf("[ninja] Received %d items for %s/%s", len(priceMap), league, itemType)
	return priceMap, nil
}
