package cli

import (
	"fmt"
	"os"
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/logging"
	"github.com/joho/godotenv"
)

// InitAppContext performs manual initialization of the App instance.
// This matches the logic in root.go's PersistentPreRunE.
func InitAppContext() error {
	_ = godotenv.Load()

	_, _ = logging.SetupLogger()
	
	var err error
	app, err = core.NewApp(cfgFile, nil)
	if err != nil {
		return fmt.Errorf("init app: %w", err)
	}

	// Default to environment URL if available
	if envURL := os.Getenv("POE_SERVER_URL"); envURL != "" {
		app.BaseURL = envURL
	}

	return nil
}
