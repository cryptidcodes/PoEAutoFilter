//go:build !windows && !linux

package gui

import (
	"fmt"
	"os"
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
)

func startGUI(app *core.App) {
	fmt.Println("GUI is not available on this platform.")
	fmt.Println("Supported platforms: Windows (lxn/walk), Linux (gotk4).")
	os.Exit(1)
}
