package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchCurrencyValues(t *testing.T) {
	// Create a mock server that returns test currency data
	mockResponse := CurrencyResponse{
		Lines: []Currency{
			{CurrencyTypeName: "Divine Orb", ChaosEquivalent: 150.0},
			{CurrencyTypeName: "Exalted Orb", ChaosEquivalent: 12.0},
		},
	}
	data, _ := json.Marshal(mockResponse)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path contains expected parameters
		if r.URL.Path != "/currencyoverview" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		league := r.URL.Query().Get("league")
		if league != "TestLeague" {
			t.Errorf("expected league TestLeague, got %s", league)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer server.Close()

	items, err := FetchCurrencyValues(server.URL, "TestLeague", "Currency")
	if err != nil {
		t.Fatalf("FetchCurrencyValues failed: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].CurrencyTypeName != "Divine Orb" {
		t.Errorf("expected Divine Orb, got %s", items[0].CurrencyTypeName)
	}
	if items[0].ChaosEquivalent != 150.0 {
		t.Errorf("expected 150.0, got %.1f", items[0].ChaosEquivalent)
	}
}

func TestFetchItemValues(t *testing.T) {
	// Create a mock server that returns test item data
	mockResponse := ItemResponse{
		Lines: []Item{
			{Name: "Gilded Scarab", ChaosValue: 25.0},
			{Name: "Polished Scarab", ChaosValue: 5.0},
		},
	}
	data, _ := json.Marshal(mockResponse)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/itemoverview" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer server.Close()

	items, err := FetchItemValues(server.URL, "TestLeague", "Scarab")
	if err != nil {
		t.Fatalf("FetchItemValues failed: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "Gilded Scarab" {
		t.Errorf("expected Gilded Scarab, got %s", items[0].Name)
	}
}

func TestFetchCurrencyValuesError(t *testing.T) {
	// Server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := FetchCurrencyValues(server.URL, "TestLeague", "Currency")
	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

func TestFetchCurrencyValuesMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid json"))
	}))
	defer server.Close()

	_, err := FetchCurrencyValues(server.URL, "TestLeague", "Currency")
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}
