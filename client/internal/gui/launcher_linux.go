//go:build linux

package gui

import (
	"log"
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/gui/linux"
)

func startGUI(app *core.App) {
	log.Println("[launcher] Launching Linux native GUI (GTK4)")
	linux.RunGUI(app)
}
