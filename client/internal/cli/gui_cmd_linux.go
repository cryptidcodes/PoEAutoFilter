//go:build linux

// gui_cmd_linux.go provides the Linux-specific GUI launch command.
// It imports the Linux GUI package and calls RunGUI to create the native
// GTK4 window using gotk4.

package cli

import (
	"log"

	linux "github.com/cryptidcodes/PoEAutoFilter/client/internal/gui/linux"
	"github.com/spf13/cobra"
)

// guiCmd launches the Linux native GUI using gotk4 (GTK4).
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical user interface",
	Long:  `Launch the Linux GTK4 native GUI for managing styles, tiers, and running the filter updater.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("[cli] Launching Linux GTK4 GUI")
		linux.RunGUI(app)
		return nil
	},
}

// runGUICmd is called when poefilter is run with no subcommand on Linux.
func runGUICmd(cmd *cobra.Command, args []string) error {
	log.Println("[cli] No subcommand specified, launching Linux GTK4 GUI")
	linux.RunGUI(app)
	return nil
}
