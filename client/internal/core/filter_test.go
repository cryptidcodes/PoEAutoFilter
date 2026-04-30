package core

import (
	"bytes"
	"os"
	"testing"
)

func TestUpdateFilterFile(t *testing.T) {
	baseFile := "test_base.filter"
	outputFile := "test_output.filter"
	defer os.Remove(baseFile)
	defer os.Remove(outputFile)

	baseContent := "Show\nBaseType == \"Mirror of Kalandra\"\n"
	err := os.WriteFile(baseFile, []byte(baseContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	newRules := "Show\nBaseType == \"Chaos Orb\"\n"
	err = UpdateFilterFile(baseFile, outputFile, newRules)
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	// Check that new rules come BEFORE base content
	newRulesIdx := bytes.Index(got, []byte(newRules))
	baseContentIdx := bytes.Index(got, []byte(baseContent))

	if newRulesIdx == -1 {
		t.Errorf("expected output to contain new rules, but it didn't")
	}
	if baseContentIdx == -1 {
		t.Errorf("expected output to contain base content, but it didn't")
	}
	if newRulesIdx > baseContentIdx {
		t.Errorf("expected new rules to be before base content, but they were after")
	}

	if !bytes.Contains(got, []byte("PoEAutoFilter - Dynamic Economy Rules")) {
		t.Errorf("expected output to contain header, but it didn't")
	}
	if !bytes.Contains(got, []byte("SECTION: Base Filter Content (Template)")) {
		t.Errorf("expected output to contain separator, but it didn't")
	}
}
func TestWriteFilterBlocks(t *testing.T) {
	cfg := Config{
		StyleLibrary: []Style{
			{Name: "High", Actions: []FilterAction{{Type: "SetFontSize", Values: []string{"45"}}}},
			{Name: "Low", Actions: []FilterAction{{Type: "SetFontSize", Values: []string{"30"}}}},
		},
		Tiers: []Tier{
			{Name: "1 Chaos", Value: 1.0, Currency: "Chaos", StyleName: "Low"},
			{Name: "10 Chaos", Value: 10.0, Currency: "Chaos", StyleName: "High"},
		},
	}

	valueMap := map[string]map[string]float64{
		"Currency": {
			"Chaos Orb": 1.0,
		},
	}

	// Mock prices
	prices := PriceTable{
		Exalted: 10.0,
		Divine:  100.0,
	}

	result := WriteFilterBlocks(cfg, valueMap, prices)

	// Check that 10 Chaos Tier comes BEFORE 1 Chaos Tier
	idx10 := bytes.Index([]byte(result), []byte("## 10 Chaos Tier ##"))
	idx1 := bytes.Index([]byte(result), []byte("## 1 Chaos Tier ##"))

	if idx10 == -1 || idx1 == -1 {
		t.Fatalf("tiers not found in result")
	}
	if idx10 > idx1 {
		t.Errorf("expected 10 Chaos Tier to be before 1 Chaos Tier")
	}

	// Check styles
	if !bytes.Contains([]byte(result), []byte("SetFontSize 45")) {
		t.Errorf("missing style SetFontSize 45")
	}
	if !bytes.Contains([]byte(result), []byte("SetFontSize 30")) {
		t.Errorf("missing style SetFontSize 30")
	}
}

func TestWriteFilterBlocksResonators(t *testing.T) {
	cfg := Config{
		StyleLibrary: []Style{
			{Name: "High", Actions: []FilterAction{{Type: "SetFontSize", Values: []string{"45"}}}},
		},
		Tiers: []Tier{
			{Name: "1 Chaos", Value: 1.0, Currency: "Chaos", StyleName: "High"},
		},
	}

	valueMap := map[string]map[string]float64{
		"Resonators": {
			"Primitive Alchemical Resonator": 2.0,
		},
	}

	prices := PriceTable{
		Exalted: 10.0,
		Divine:  100.0,
	}

	result := WriteFilterBlocks(cfg, valueMap, prices)

	if !bytes.Contains([]byte(result), []byte("Primitive Alchemical Resonator")) {
		t.Errorf("expected Resonator filter blocks, got none. Output:\n%s", result)
	}
	if !bytes.Contains([]byte(result), []byte("Show")) {
		t.Errorf("expected Show block for Resonator")
	}
}
