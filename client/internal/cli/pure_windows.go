//go:build windows

package cli

import (
	"log"
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/gui/windows"
)

// RunGUIPure initializes the app context and launches the Windows GUI directly.
func RunGUIPure() {
	if err := InitAppContext(); err != nil {
		return
	}

	app.Log("[CLI] App context initialized. Launching GUI...\n")
	log.Println("[cli] Launching Windows GUI (Pure)")
	windows.RunGUI(app)
}
