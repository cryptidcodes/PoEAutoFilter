package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configCmd implements the `poefilter config` subcommand for reading and modifying
// configuration values (league, filepaths) from the command line.
//
// Usage:
//
//	poefilter config get league         → prints current league
//	poefilter config set league Settlers → sets league to "Settlers"
//	poefilter config get filepath       → prints output filter path
//	poefilter config set filepath /path → sets output filter path
//	poefilter config get basepath       → prints base filter path
//	poefilter config set basepath /path → sets base filter path
//	poefilter config show               → prints all config values
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or modify configuration values",
	Long: `View or modify configuration values such as league, file paths, and override text.

Subcommands: get, set, show`,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long:  `Get a configuration value. Valid keys: league, filepath, basepath, override`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		switch key {
		case "league":
			fmt.Println(app.Config.League)
		case "filepath":
			fmt.Println(app.Config.FilePath)
		case "basepath":
			fmt.Println(app.Config.BaseFilePath)
		case "override":
			fmt.Println(app.Config.Override)
		default:
			return fmt.Errorf("unknown config key: %q (valid: league, filepath, basepath, override)", key)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration value. Valid keys: league, filepath, basepath, override`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		switch key {
		case "league":
			app.Config.League = value
		case "filepath":
			app.Config.FilePath = value
		case "basepath":
			app.Config.BaseFilePath = value
		case "override":
			app.Config.Override = value
		default:
			return fmt.Errorf("unknown config key: %q (valid: league, filepath, basepath, override)", key)
		}

		app.UpdateConfig(app.Config)
		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%-15s %s\n", "League:", app.Config.League)
		fmt.Printf("%-15s %s\n", "Filter Path:", app.Config.FilePath)
		fmt.Printf("%-15s %s\n", "Base Path:", app.Config.BaseFilePath)
		fmt.Printf("%-15s %d styles\n", "Styles:", len(app.Config.StyleLibrary))
		fmt.Printf("%-15s %d tiers\n", "Tiers:", len(app.Config.Tiers))
		if app.Config.Override != "" {
			fmt.Printf("%-15s (set, %d chars)\n", "Override:", len(app.Config.Override))
		} else {
			fmt.Printf("%-15s (not set)\n", "Override:")
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
}
