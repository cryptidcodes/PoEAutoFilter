package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// runCmd implements the `poefilter run` subcommand for one-shot filter updates.
// It fetches current prices, generates filter rules, writes the output file, and exits.
// This is the headless equivalent of clicking "Start AutoFilter" and waiting for one cycle.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch prices and update the filter file (one-shot)",
	Long: `Fetch current economy data from poe.ninja (or configured server),
generate filter rules based on your tiers and styles, and write the output filter file.
This runs once and exits. Use 'watch' for continuous updates.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("[cli] Running one-shot filter update")

		// Set up a log function for stdout output
		app.LogFunc = func(msg string) {
			fmt.Print(msg)
		}

		ctx := context.Background()
		app.ProcessFilterUpdate(ctx)

		return nil
	},
}
