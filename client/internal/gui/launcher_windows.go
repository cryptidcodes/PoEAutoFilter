//go:build windows

package gui

import (
	"log"
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/gui/windows"
)

func startGUI(app *core.App) {
	log.Println("[launcher] Launching Windows native GUI (lxn/walk)")
	windows.RunGUI(app)
}
