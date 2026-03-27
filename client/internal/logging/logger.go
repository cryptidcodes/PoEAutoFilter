// Package logging provides cross-platform logger setup for PoEAutoFilter.
// It configures Go's standard logger to write to both stdout and a debug.log file.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// SetupLogger configures the global logger to write to a debug.log file and stdout.
// Returns the log file handle which should be deferred for cleanup.
// The log file is created next to the executable for easy discovery.
func SetupLogger() (*os.File, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exPath := filepath.Dir(ex)
	logPath := filepath.Join(exPath, "debug.log")

	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("error opening log file: %v", err)
	}

	// Write to both file and stdout
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("=== Session Started: %s ===", time.Now().Format(time.RFC3339))
	return f, nil
}
