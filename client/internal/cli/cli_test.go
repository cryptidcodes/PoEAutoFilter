package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
)

// setupTestApp creates a temporary config file and initializes the global app
// variable for testing CLI commands.
func setupTestApp(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := core.Config{
		League:   "TestLeague",
		FilePath: "/tmp/test_output.filter",
		StyleLibrary: []core.Style{
			{Name: "High", Actions: []core.FilterAction{{Type: "SetFontSize", Values: []string{"45"}}}},
			{Name: "Low", Actions: []core.FilterAction{{Type: "SetFontSize", Values: []string{"30"}}}},
		},
		Tiers: []core.Tier{
			{Name: "10c", Value: 10.0, Currency: "Chaos", StyleName: "High"},
			{Name: "1c", Value: 1.0, Currency: "Chaos", StyleName: "Low"},
		},
	}

	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(cfgPath, data, 0644)

	var err error
	cfgFile = cfgPath
	app, err = core.NewApp(cfgPath, nil)
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}

	return cfgPath
}

func TestConfigGetLeague(t *testing.T) {
	setupTestApp(t)

	// Test that we can read the league from the loaded config
	if app.Config.League != "TestLeague" {
		t.Errorf("expected league TestLeague, got %s", app.Config.League)
	}
}

func TestConfigSetLeague(t *testing.T) {
	cfgPath := setupTestApp(t)

	app.Config.League = "NewLeague"
	app.UpdateConfig(app.Config)

	// Verify persistence
	loaded, err := core.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if loaded.League != "NewLeague" {
		t.Errorf("expected league NewLeague after set, got %s", loaded.League)
	}
}

func TestStylesList(t *testing.T) {
	setupTestApp(t)

	result := core.ListStyles(app.Config)
	if !strings.Contains(result, "High") {
		t.Error("expected styles list to contain 'High'")
	}
	if !strings.Contains(result, "Low") {
		t.Error("expected styles list to contain 'Low'")
	}
}

func TestStylesAddAndRemove(t *testing.T) {
	setupTestApp(t)

	// Add
	newStyle := core.Style{Name: "Medium", Actions: []core.FilterAction{{Type: "SetFontSize", Values: []string{"36"}}}}
	err := core.AddStyle(&app.Config, newStyle)
	if err != nil {
		t.Fatalf("AddStyle failed: %v", err)
	}
	if len(app.Config.StyleLibrary) != 3 {
		t.Errorf("expected 3 styles, got %d", len(app.Config.StyleLibrary))
	}

	// Remove
	err = core.RemoveStyle(&app.Config, "Medium")
	if err != nil {
		t.Fatalf("RemoveStyle failed: %v", err)
	}
	if len(app.Config.StyleLibrary) != 2 {
		t.Errorf("expected 2 styles, got %d", len(app.Config.StyleLibrary))
	}
}

func TestTiersList(t *testing.T) {
	setupTestApp(t)

	result := core.ListTiers(app.Config)
	if !strings.Contains(result, "10c") {
		t.Error("expected tiers list to contain '10c'")
	}
	if !strings.Contains(result, "1c") {
		t.Error("expected tiers list to contain '1c'")
	}
}

func TestTiersAddAndRemove(t *testing.T) {
	setupTestApp(t)

	// Add
	newTier := core.Tier{Name: "50c", Value: 50.0, Currency: "Chaos", StyleName: "High"}
	err := core.AddTier(&app.Config, newTier)
	if err != nil {
		t.Fatalf("AddTier failed: %v", err)
	}
	if len(app.Config.Tiers) != 3 {
		t.Errorf("expected 3 tiers, got %d", len(app.Config.Tiers))
	}

	// Remove
	err = core.RemoveTier(&app.Config, "50c")
	if err != nil {
		t.Fatalf("RemoveTier failed: %v", err)
	}
	if len(app.Config.Tiers) != 2 {
		t.Errorf("expected 2 tiers, got %d", len(app.Config.Tiers))
	}
}
