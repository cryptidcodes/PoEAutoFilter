//go:build windows

// gui_cmd_windows.go provides the Windows-specific GUI launch command.
// It imports the Windows GUI package and calls RunGUI to create the native
// Win32 window using lxn/walk.

package cli

import (
	"log"

	windows "github.com/cryptidcodes/PoEAutoFilter/client/internal/gui/windows"
	"github.com/spf13/cobra"
)

// guiCmd launches the Windows native GUI using lxn/walk.
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical user interface",
	Long:  `Launch the Windows native GUI for managing styles, tiers, and running the filter updater.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("[cli] Launching Windows GUI")
		windows.RunGUI(app)
		return nil
	},
}

// runGUICmd is called when poefilter is run with no subcommand on Windows.
func runGUICmd(cmd *cobra.Command, args []string) error {
	log.Println("[cli] No subcommand specified, launching Windows GUI")
	windows.RunGUI(app)
	return nil
}
