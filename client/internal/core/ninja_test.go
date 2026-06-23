package core

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestFetchPrices(t *testing.T) {
	// Set mock environment variables for the test
	os.Setenv("API_URL", "/api/economy")
	os.Setenv("PRICE_SOURCE_EXCHANGE", "/exchange")
	os.Setenv("CURRENT_PRICE", "/current/overview")
	defer func() {
		os.Unsetenv("API_URL")
		os.Unsetenv("PRICE_SOURCE_EXCHANGE")
		os.Unsetenv("CURRENT_PRICE")
	}()

	mockResponse := NinjaResponse{
		Core: NinjaCore{
			Rates:     map[string]float64{"exalted": 347.1, "chaos": 8.30},
			Primary:   "divine",
			Secondary: "chaos",
		},
		Items: []NinjaItem{
			{ID: "divine", Name: "Divine Orb"},
			{ID: "exalted", Name: "Exalted Orb"},
		},
		Lines: []NinjaLine{
			{ID: "divine", PrimaryValue: 1.0},    // 1 divine
			{ID: "exalted", PrimaryValue: 0.003}, // 0.003 divines
		},
	}
	data, _ := json.Marshal(mockResponse)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPathPoE1 := "/poe1/api/economy/exchange/current/overview"
		expectedPathPoE2 := "/poe2/api/economy/exchange/current/overview"
		if r.URL.Path != expectedPathPoE1 && r.URL.Path != expectedPathPoE2 {
			t.Errorf("unexpected path: got %s", r.URL.Path)
		}
		league := r.URL.Query().Get("league")
		if league != "TestLeague" {
			t.Errorf("expected league TestLeague, got %s", league)
		}
		itemType := r.URL.Query().Get("type")
		if itemType != "Currency" {
			t.Errorf("expected type Currency, got %s", itemType)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer server.Close()

	// poe1: should convert using chaos rate (8.30)
	priceMap, err := FetchPrices(server.URL, "poe1", "TestLeague", "Currency")
	if err != nil {
		t.Fatalf("FetchPrices failed: %v", err)
	}

	if len(priceMap) != 2 {
		t.Fatalf("expected 2 items, got %d", len(priceMap))
	}
	// Divine Orb: 1.0 * 8.30 = 8.30 chaos
	expectedDivine := 1.0 * 8.30
	if math.Abs(priceMap["Divine Orb"]-expectedDivine) > 0.01 {
		t.Errorf("expected %.2f chaos for Divine Orb, got %.2f", expectedDivine, priceMap["Divine Orb"])
	}
	// Exalted Orb: 0.003 * 8.30 = 0.0249 chaos
	expectedExalted := 0.003 * 8.30
	if math.Abs(priceMap["Exalted Orb"]-expectedExalted) > 0.001 {
		t.Errorf("expected %.4f chaos for Exalted Orb, got %.4f", expectedExalted, priceMap["Exalted Orb"])
	}

	// poe2: should convert using exalted rate (347.1)
	priceMapPoE2, err := FetchPrices(server.URL, "poe2", "TestLeague", "Currency")
	if err != nil {
		t.Fatalf("FetchPrices for poe2 failed: %v", err)
	}
	if len(priceMapPoE2) != 2 {
		t.Fatalf("expected 2 items for poe2, got %d", len(priceMapPoE2))
	}
	// Divine Orb: 1.0 * 347.1 = 347.1 exalted
	expectedDivinePoE2 := 1.0 * 347.1
	if math.Abs(priceMapPoE2["Divine Orb"]-expectedDivinePoE2) > 0.01 {
		t.Errorf("expected %.2f exalted for Divine Orb (poe2), got %.2f", expectedDivinePoE2, priceMapPoE2["Divine Orb"])
	}
	// Exalted Orb: 0.003 * 347.1 = 1.0413 exalted
	expectedExaltedPoE2 := 0.003 * 347.1
	if math.Abs(priceMapPoE2["Exalted Orb"]-expectedExaltedPoE2) > 0.01 {
		t.Errorf("expected %.4f exalted for Exalted Orb (poe2), got %.4f", expectedExaltedPoE2, priceMapPoE2["Exalted Orb"])
	}
}

func TestFetchPricesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := FetchPrices(server.URL, "poe1", "TestLeague", "Currency")
	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

func TestFetchPricesMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid json"))
	}))
	defer server.Close()

	_, err := FetchPrices(server.URL, "poe1", "TestLeague", "Currency")
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestFetchPricesMissingRate(t *testing.T) {
	os.Setenv("API_URL", "/api/economy")
	os.Setenv("PRICE_SOURCE_EXCHANGE", "/exchange")
	os.Setenv("CURRENT_PRICE", "/current/overview")
	defer func() {
		os.Unsetenv("API_URL")
		os.Unsetenv("PRICE_SOURCE_EXCHANGE")
		os.Unsetenv("CURRENT_PRICE")
	}()

	// Response with no rates — should fall back to rate=1.0
	mockResponse := NinjaResponse{
		Core: NinjaCore{
			Primary: "divine",
		},
		Items: []NinjaItem{{ID: "divine", Name: "Divine Orb"}},
		Lines: []NinjaLine{{ID: "divine", PrimaryValue: 1.0}},
	}
	data, _ := json.Marshal(mockResponse)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer server.Close()

	priceMap, err := FetchPrices(server.URL, "poe1", "TestLeague", "Currency")
	if err != nil {
		t.Fatalf("FetchPrices with missing rate failed: %v", err)
	}
	// With no rate, should use primaryValue as-is (rate=1.0)
	if priceMap["Divine Orb"] != 1.0 {
		t.Errorf("expected 1.0 (fallback), got %.4f", priceMap["Divine Orb"])
	}
}

func TestFetchPricesLeagueWithSpaces(t *testing.T) {
	os.Setenv("API_URL", "/api/economy")
	os.Setenv("PRICE_SOURCE_EXCHANGE", "/exchange")
	os.Setenv("CURRENT_PRICE", "/current/overview")
	defer func() {
		os.Unsetenv("API_URL")
		os.Unsetenv("PRICE_SOURCE_EXCHANGE")
		os.Unsetenv("CURRENT_PRICE")
	}()

	mockResponse := NinjaResponse{
		Core: NinjaCore{
			Rates:     map[string]float64{"exalted": 347.1, "chaos": 8.30},
			Primary:   "divine",
			Secondary: "chaos",
		},
		Items: []NinjaItem{{ID: "divine", Name: "Divine Orb"}},
		Lines: []NinjaLine{{ID: "divine", PrimaryValue: 1.0}},
	}
	data, _ := json.Marshal(mockResponse)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		league := r.URL.Query().Get("league")
		if league != "Runes of Aldur" {
			t.Errorf("expected league 'Runes of Aldur', got '%s'", league)
		}
		// Verify the raw query string uses '+' for spaces
		rawQuery := r.URL.RawQuery
		if !strings.Contains(rawQuery, "league=Runes+of+Aldur") && !strings.Contains(rawQuery, "league=Runes%20of%20Aldur") {
			t.Errorf("expected URL-encoded league in raw query, got: %s", rawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer server.Close()

	priceMap, err := FetchPrices(server.URL, "poe1", "Runes of Aldur", "Currency")
	if err != nil {
		t.Fatalf("FetchPrices with spaced league failed: %v", err)
	}
	if len(priceMap) != 1 {
		t.Fatalf("expected 1 item, got %d", len(priceMap))
	}
	// 1.0 * 8.30 = 8.30 chaos
	expected := 1.0 * 8.30
	if math.Abs(priceMap["Divine Orb"]-expected) > 0.01 {
		t.Errorf("expected %.2f for Divine Orb, got %.2f", expected, priceMap["Divine Orb"])
	}
}
