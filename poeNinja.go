package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

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
