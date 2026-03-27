// Package cli implements the cobra-based command-line interface for PoEAutoFilter.
// The root command provides global flags (--config, --verbose, --server-url) and
// dispatches to subcommands. When run with no subcommand, it defaults to launching the GUI.
package cli

import (
	"log"
	"os"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
	"github.com/spf13/cobra"
)

var (
	// Global flags shared across all subcommands
	cfgFile   string
	verbose   bool
	serverURL string

	// app is the shared application instance, initialized in PersistentPreRunE
	app *core.App
)

// rootCmd is the base command for the PoEAutoFilter CLI.
// Running without a subcommand defaults to the GUI.
var rootCmd = &cobra.Command{
	Use:   "poefilter",
	Short: "PoEAutoFilter — dynamic economy-based item filter generator for Path of Exile",
	Long: `PoEAutoFilter fetches live economy data from poe.ninja and generates
dynamic item filter rules based on configurable price tiers and visual styles.

Run without arguments to launch the GUI, or use subcommands for headless operation.`,
	// PersistentPreRunE initializes the app for ALL subcommands before they execute.
	// This ensures config is loaded and logging is set up once, regardless of which
	// subcommand the user invokes.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Use the common initialization logic
		if err := InitAppContext(); err != nil {
			return err
		}

		// Apply CLI-specific overrides from flags
		if serverURL != "" {
			app.BaseURL = serverURL
			log.Printf("[cli] Using server URL from flag: %s", serverURL)
		}

		if verbose {
			log.Println("[cli] Verbose mode enabled")
		}

		return nil
	},
	// Default action when no subcommand is given: launch GUI
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGUICmd(cmd, args)
	},
}

// Execute runs the root command. Called from main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.json",
		"Path to the configuration file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", "",
		"Custom server URL for price data (default: poe.ninja direct)")

	// Register all subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(stylesCmd)
	rootCmd.AddCommand(tiersCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(guiCmd)
}
