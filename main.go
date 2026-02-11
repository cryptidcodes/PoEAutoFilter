package main

import (
	"fmt"
	"log"
)

// main is the entry point.
func main() {
	// Setup Logger
	f, err := SetupLogger()
	if err != nil {
		// Fallback to basic print if logger fails
		fmt.Printf("Failed to setup logger: %v\n", err)
	} else {
		defer f.Close()
	}

	// Initialize App
	app, err := NewApp("config.json", nil)
	if err != nil {
		log.Printf("Fatal error initializing app: %v", err)
		// Show error in a blocking way if possible, or just exit
		// Since GUI isn't up, we just log and exit
		return
	}

	// Launch GUI
	RunGUI(app)
}
