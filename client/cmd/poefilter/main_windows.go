//go:build windows

package main

import (
	"os"
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/cli"
)

func main() {
	// If no arguments are provided on Windows, launch the GUI directly
	// to avoid Cobra's console-centric help/usage from flashing.
	if len(os.Args) == 1 {
		cli.RunGUIPure()
		return
	}
	cli.Execute()
}
