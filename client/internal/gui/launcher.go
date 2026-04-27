package gui

import (
	"log"
	"os"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/logging"
	"github.com/joho/godotenv"
)

// RunApp initializes the application and launches the platform-specific GUI.
func RunApp() {
	// Load environment variables if .env exists
	_ = godotenv.Load()

	// Setup logging
	_, _ = logging.SetupLogger()
	log.Println("[launcher] Initializing application...")

	// Initialize core app with default config
	app, err := core.NewApp("config.json", nil)
	if err != nil {
		log.Fatalf("[launcher] Fatal: Failed to initialize core app: %v", err)
	}

	// Apply environment overrides
	if envURL := os.Getenv("AUTOFILTER_SERVER_URL"); envURL != "" {
		app.BaseURL = envURL
		log.Printf("[launcher] Using server URL from environment: %s", envURL)
	}

	// Start the platform-specific GUI
	startGUI(app)
}
