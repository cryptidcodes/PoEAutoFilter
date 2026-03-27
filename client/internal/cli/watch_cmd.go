package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var watchInterval time.Duration

// watchCmd implements the `poefilter watch` subcommand for continuous filter updates.
// It runs the same logic as `run` but repeats on a configurable interval (default 1 hour).
// Handles SIGINT/SIGTERM for graceful shutdown.
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Continuously fetch prices and update the filter (daemon mode)",
	Long: `Continuously fetch economy data and regenerate filter rules on a timer.
Default interval is 1 hour. Use --interval to customize.

Press Ctrl+C or send SIGTERM to stop gracefully.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("[cli] Starting watch mode with interval: %v", watchInterval)

		// Set up a log function for stdout output
		app.LogFunc = func(msg string) {
			fmt.Print(msg)
		}

		// Create cancellable context for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle OS signals for graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigCh
			log.Printf("[cli] Received signal %v, shutting down...", sig)
			fmt.Printf("\nReceived %v, shutting down gracefully...\n", sig)
			cancel()
		}()

		// Run immediately once
		app.ProcessFilterUpdate(ctx)

		// Then loop on ticker
		ticker := time.NewTicker(watchInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Watch mode stopped.")
				return nil
			case <-ticker.C:
				app.ProcessFilterUpdate(ctx)
			}
		}
	},
}

func init() {
	watchCmd.Flags().DurationVar(&watchInterval, "interval", 1*time.Hour,
		"Interval between price checks (e.g., 30m, 1h, 2h)")
}
