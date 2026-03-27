package core

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Currency represents a single currency item from the poe.ninja currencyoverview API.
// The ChaosEquivalent field is the primary value used for filter generation.
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

// Item represents a single item from the poe.ninja itemoverview API.
// Used for categories like Scarabs, Fossils, Essences, etc.
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

// CurrencyResponse is the top-level JSON structure returned by the poe.ninja currencyoverview API.
type CurrencyResponse struct {
	Lines []Currency `json:"lines"`
}

// ItemResponse is the top-level JSON structure returned by the poe.ninja itemoverview API.
type ItemResponse struct {
	Lines []Item `json:"lines"`
}

// DefaultBaseURL is the default server URL used by the app.
// This is a variable so it can be overridden at build-time using:
// -ldflags="-X 'github.com/cryptidcodes/PoEAutoFilter/client/internal/core.DefaultBaseURL=https://your-domain.com'"
var DefaultBaseURL = "https://poe.ninja/api/data"

// FetchCurrencyValues retrieves currency data from the poe.ninja currencyoverview endpoint.
// The baseURL parameter allows redirecting requests to a custom server (for edge deployments).
// Returns a slice of Currency objects or an error if the request fails.
func FetchCurrencyValues(baseURL, league, itemType string) ([]Currency, error) {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	url := fmt.Sprintf("%s/currencyoverview?league=%s&type=%s", baseURL, league, itemType)
	log.Printf("[ninja] Fetching currency: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[ninja] HTTP error fetching currency: %v", err)
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

	var data CurrencyResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("[ninja] Error parsing currency JSON: %v", err)
		return nil, err
	}

	log.Printf("[ninja] Received %d currency items for %s/%s", len(data.Lines), league, itemType)
	return data.Lines, nil
}

// FetchItemValues retrieves item data (scarabs, fossils, essences, etc.) from the
// poe.ninja itemoverview endpoint. The baseURL parameter allows redirecting requests
// to a custom server (for edge deployments).
// Returns a slice of Item objects or an error if the request fails.
func FetchItemValues(baseURL, league, itemType string) ([]Item, error) {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	url := fmt.Sprintf("%s/itemoverview?league=%s&type=%s", baseURL, league, itemType)
	log.Printf("[ninja] Fetching items: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[ninja] HTTP error fetching items: %v", err)
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

	var data ItemResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("[ninja] Error parsing item JSON: %v", err)
		return nil, err
	}

	log.Printf("[ninja] Received %d items for %s/%s", len(data.Lines), league, itemType)
	return data.Lines, nil
}
