package cli

import (
	"fmt"
	"strings"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
	"github.com/spf13/cobra"
)

// stylesCmd implements the `poefilter styles` subcommand for listing, adding,
// editing, and removing filter styles from the command line.
//
// Usage:
//
//	poefilter styles list                  → prints all styles with action previews
//	poefilter styles add --name "High"     → creates a new style
//	poefilter styles remove --name "High"  → removes a style by name
//	poefilter styles show --name "High"    → shows detailed style info
var stylesCmd = &cobra.Command{
	Use:   "styles",
	Short: "Manage filter styles",
	Long:  `List, add, show, or remove filter styles from the style library.`,
}

var stylesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all styles in the style library",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(core.ListStyles(app.Config))
		return nil
	},
}

var (
	styleAddName    string
	styleAddActions string
)

var stylesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new style to the library",
	Long: `Add a new style to the library. Specify actions as semicolon-separated filter lines.
Example: poefilter styles add --name "High Value" --actions "SetFontSize 45;SetTextColor 255 0 0 255"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if styleAddName == "" {
			return fmt.Errorf("--name is required")
		}

		style := core.Style{Name: styleAddName}

		if styleAddActions != "" {
			// Parse semicolon-separated action strings
			for _, actionStr := range strings.Split(styleAddActions, ";") {
				actionStr = strings.TrimSpace(actionStr)
				if actionStr == "" {
					continue
				}
				parts := strings.Fields(actionStr)
				if len(parts) == 0 {
					continue
				}
				style.Actions = append(style.Actions, core.FilterAction{
					Type:   parts[0],
					Values: parts[1:],
				})
			}
		}

		if err := core.AddStyle(&app.Config, style); err != nil {
			return err
		}

		app.UpdateConfig(app.Config)
		fmt.Printf("Added style %q with %d actions.\n", style.Name, len(style.Actions))
		return nil
	},
}

var styleRemoveName string

var stylesRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a style from the library",
	RunE: func(cmd *cobra.Command, args []string) error {
		if styleRemoveName == "" {
			return fmt.Errorf("--name is required")
		}

		if err := core.RemoveStyle(&app.Config, styleRemoveName); err != nil {
			return err
		}

		app.UpdateConfig(app.Config)
		fmt.Printf("Removed style %q.\n", styleRemoveName)
		return nil
	},
}

var styleShowName string

var stylesShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show detailed information about a style",
	RunE: func(cmd *cobra.Command, args []string) error {
		if styleShowName == "" {
			return fmt.Errorf("--name is required")
		}

		style := core.FindStyleByName(&app.Config, styleShowName)
		if style == nil {
			return fmt.Errorf("style %q not found", styleShowName)
		}

		fmt.Printf("Style: %s\n", style.Name)
		fmt.Printf("Actions (%d):\n", len(style.Actions))
		for i, action := range style.Actions {
			fmt.Printf("  %d. %s\n", i+1, action.ToFilterLine())
		}
		fmt.Println("\nFilter Preview:")
		fmt.Println(style.ToFilterLines())
		return nil
	},
}

func init() {
	stylesAddCmd.Flags().StringVar(&styleAddName, "name", "", "Name for the new style (required)")
	stylesAddCmd.Flags().StringVar(&styleAddActions, "actions", "",
		"Semicolon-separated filter actions (e.g., \"SetFontSize 45;SetTextColor 255 0 0 255\")")

	stylesRemoveCmd.Flags().StringVar(&styleRemoveName, "name", "", "Name of the style to remove (required)")

	stylesShowCmd.Flags().StringVar(&styleShowName, "name", "", "Name of the style to show (required)")

	stylesCmd.AddCommand(stylesListCmd)
	stylesCmd.AddCommand(stylesAddCmd)
	stylesCmd.AddCommand(stylesRemoveCmd)
	stylesCmd.AddCommand(stylesShowCmd)
}
