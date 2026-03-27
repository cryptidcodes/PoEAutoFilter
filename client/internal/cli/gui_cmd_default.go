//go:build !windows && !linux

// gui_cmd_default.go provides a fallback GUI command for unsupported platforms.
// On Windows and Linux, this file is excluded and replaced by platform-specific
// implementations that launch the native GUI (lxn/walk on Windows, gotk4 on Linux).

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// guiCmd launches the platform-specific GUI, or prints a message on unsupported platforms.
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical user interface",
	Long:  `Launch the platform-specific GUI for managing styles, tiers, and running the filter updater.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("GUI is not available on this platform.")
		fmt.Println("Supported platforms: Windows (lxn/walk), Linux (gotk4).")
		fmt.Println("Use CLI subcommands instead: styles, tiers, config, run, watch")
		return nil
	},
}

// runGUICmd is the function called when poefilter is run with no subcommand.
// On unsupported platforms, it falls back to printing help.
func runGUICmd(cmd *cobra.Command, args []string) error {
	fmt.Println("GUI is not available on this platform.")
	fmt.Println("Use --help to see available CLI commands.")
	return nil
}
