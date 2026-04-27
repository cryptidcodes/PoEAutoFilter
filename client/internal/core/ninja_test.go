package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
		Items: []NinjaItem{
			{ID: "divine", Name: "Divine Orb"},
			{ID: "exalted", Name: "Exalted Orb"},
		},
		Lines: []NinjaLine{
			{ID: "divine", PrimaryValue: 150.0},
			{ID: "exalted", PrimaryValue: 12.0},
		},
	}
	data, _ := json.Marshal(mockResponse)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/poe1/api/economy/exchange/current/overview"
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected path: expected %s, got %s", expectedPath, r.URL.Path)
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

	priceMap, err := FetchPrices(server.URL, "TestLeague", "Currency")
	if err != nil {
		t.Fatalf("FetchPrices failed: %v", err)
	}

	if len(priceMap) != 2 {
		t.Fatalf("expected 2 items, got %d", len(priceMap))
	}
	if priceMap["Divine Orb"] != 150.0 {
		t.Errorf("expected 150.0 for Divine Orb, got %.1f", priceMap["Divine Orb"])
	}
	if priceMap["Exalted Orb"] != 12.0 {
		t.Errorf("expected 12.0 for Exalted Orb, got %.1f", priceMap["Exalted Orb"])
	}
}

func TestFetchPricesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := FetchPrices(server.URL, "TestLeague", "Currency")
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

	_, err := FetchPrices(server.URL, "TestLeague", "Currency")
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}
