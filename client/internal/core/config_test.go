package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigJSON(t *testing.T) {
	// Create a temp directory for test config
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := Config{
		League:   "TestLeague",
		FilePath: "/tmp/output.filter",
		StyleLibrary: []Style{
			{Name: "TestStyle", Actions: []FilterAction{{Type: "SetFontSize", Values: []string{"40"}}}},
		},
		Tiers: []Tier{
			{Name: "TestTier", Value: 5.0, Currency: "Chaos", StyleName: "TestStyle"},
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.League != "TestLeague" {
		t.Errorf("expected league TestLeague, got %s", loaded.League)
	}
	if len(loaded.StyleLibrary) != 1 {
		t.Errorf("expected 1 style, got %d", len(loaded.StyleLibrary))
	}
	if len(loaded.Tiers) != 1 {
		t.Errorf("expected 1 tier, got %d", len(loaded.Tiers))
	}
}

func TestLoadConfigMissing(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "nonexistent.json")

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig should not error on missing file, got: %v", err)
	}
	if cfg.League != "Standard" {
		t.Errorf("expected default league Standard, got %s", cfg.League)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "roundtrip.json")

	original := Config{
		League:       "Settlers",
		FilePath:     "/home/user/filter.filter",
		BaseFilePath: "/home/user/base.filter",
		StyleLibrary: []Style{
			{Name: "Divine", Actions: []FilterAction{
				{Type: "SetFontSize", Values: []string{"45"}},
				{Type: "SetTextColor", Values: []string{"255", "0", "0", "255"}},
			}},
		},
		Tiers: []Tier{
			{Name: "1 Divine", Value: 1.0, Currency: "Divine", StyleName: "Divine"},
		},
	}

	if err := SaveConfig(original, cfgPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.League != original.League {
		t.Errorf("league mismatch: got %s, want %s", loaded.League, original.League)
	}
	if len(loaded.StyleLibrary) != 1 || loaded.StyleLibrary[0].Name != "Divine" {
		t.Errorf("style mismatch after round trip")
	}
	if len(loaded.Tiers) != 1 || loaded.Tiers[0].Name != "1 Divine" {
		t.Errorf("tier mismatch after round trip")
	}
}

func TestAddStyle(t *testing.T) {
	cfg := Config{}
	style := Style{Name: "New", Actions: []FilterAction{{Type: "SetFontSize", Values: []string{"32"}}}}

	if err := AddStyle(&cfg, style); err != nil {
		t.Fatalf("AddStyle failed: %v", err)
	}
	if len(cfg.StyleLibrary) != 1 {
		t.Fatalf("expected 1 style, got %d", len(cfg.StyleLibrary))
	}

	// Duplicate should fail
	if err := AddStyle(&cfg, style); err == nil {
		t.Error("expected error for duplicate style, got nil")
	}
}

func TestRemoveStyle(t *testing.T) {
	cfg := Config{
		StyleLibrary: []Style{
			{Name: "A"}, {Name: "B"}, {Name: "C"},
		},
	}

	if err := RemoveStyle(&cfg, "B"); err != nil {
		t.Fatalf("RemoveStyle failed: %v", err)
	}
	if len(cfg.StyleLibrary) != 2 {
		t.Errorf("expected 2 styles, got %d", len(cfg.StyleLibrary))
	}
	if cfg.StyleLibrary[0].Name != "A" || cfg.StyleLibrary[1].Name != "C" {
		t.Errorf("unexpected remaining styles: %v", cfg.StyleLibrary)
	}

	// Not found
	if err := RemoveStyle(&cfg, "Z"); err == nil {
		t.Error("expected error for missing style, got nil")
	}
}

func TestAddTier(t *testing.T) {
	cfg := Config{}
	tier := Tier{Name: "10c", Value: 10.0, Currency: "Chaos", StyleName: "High"}

	if err := AddTier(&cfg, tier); err != nil {
		t.Fatalf("AddTier failed: %v", err)
	}
	if len(cfg.Tiers) != 1 {
		t.Fatalf("expected 1 tier, got %d", len(cfg.Tiers))
	}

	// Duplicate
	if err := AddTier(&cfg, tier); err == nil {
		t.Error("expected error for duplicate tier, got nil")
	}
}

func TestRemoveTier(t *testing.T) {
	cfg := Config{
		Tiers: []Tier{
			{Name: "A"}, {Name: "B"},
		},
	}

	if err := RemoveTier(&cfg, "A"); err != nil {
		t.Fatalf("RemoveTier failed: %v", err)
	}
	if len(cfg.Tiers) != 1 {
		t.Errorf("expected 1 tier, got %d", len(cfg.Tiers))
	}
	if cfg.Tiers[0].Name != "B" {
		t.Errorf("expected remaining tier B, got %s", cfg.Tiers[0].Name)
	}
}

func TestListStyles(t *testing.T) {
	cfg := Config{
		StyleLibrary: []Style{
			{Name: "High", Actions: []FilterAction{{Type: "SetFontSize", Values: []string{"45"}}}},
		},
	}
	result := ListStyles(cfg)
	if result == "" {
		t.Error("expected non-empty output")
	}
	if !contains(result, "High") {
		t.Error("expected output to contain 'High'")
	}
}

func TestListTiers(t *testing.T) {
	cfg := Config{
		Tiers: []Tier{
			{Name: "10c", Value: 10.0, Currency: "Chaos", StyleName: "High"},
		},
	}
	result := ListTiers(cfg)
	if result == "" {
		t.Error("expected non-empty output")
	}
	if !contains(result, "10c") {
		t.Error("expected output to contain '10c'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
