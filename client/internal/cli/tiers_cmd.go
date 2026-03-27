package cli

import (
	"fmt"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
	"github.com/spf13/cobra"
)

// tiersCmd implements the `poefilter tiers` subcommand for listing, adding,
// editing, and removing value tiers from the command line.
//
// Usage:
//
//	poefilter tiers list                                            → prints all tiers
//	poefilter tiers add --name "50c" --value 50 --currency Chaos --style "High"
//	poefilter tiers edit --name "50c" --value 100
//	poefilter tiers remove --name "50c"
var tiersCmd = &cobra.Command{
	Use:   "tiers",
	Short: "Manage value tiers",
	Long:  `List, add, edit, or remove value tiers that map price thresholds to visual styles.`,
}

var tiersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all value tiers",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(core.ListTiers(app.Config))
		return nil
	},
}

var (
	tierAddName     string
	tierAddValue    float64
	tierAddCurrency string
	tierAddStyle    string
)

var tiersAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new value tier",
	RunE: func(cmd *cobra.Command, args []string) error {
		if tierAddName == "" {
			return fmt.Errorf("--name is required")
		}
		if tierAddCurrency == "" {
			tierAddCurrency = "Chaos"
		}
		if tierAddStyle == "" {
			return fmt.Errorf("--style is required (name of an existing style)")
		}

		// Validate that the style exists
		if core.FindStyleByName(&app.Config, tierAddStyle) == nil {
			return fmt.Errorf("style %q not found in style library", tierAddStyle)
		}

		tier := core.Tier{
			Name:      tierAddName,
			Value:     tierAddValue,
			Currency:  tierAddCurrency,
			StyleName: tierAddStyle,
		}

		if err := core.AddTier(&app.Config, tier); err != nil {
			return err
		}

		app.UpdateConfig(app.Config)
		fmt.Printf("Added tier %q (%.2f %s → style %q).\n",
			tier.Name, tier.Value, tier.Currency, tier.StyleName)
		return nil
	},
}

var (
	tierEditName     string
	tierEditValue    float64
	tierEditCurrency string
	tierEditStyle    string
)

var tiersEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an existing value tier",
	Long:  `Edit an existing value tier. Only specified flags will be updated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if tierEditName == "" {
			return fmt.Errorf("--name is required to identify the tier to edit")
		}

		tier := core.FindTierByName(&app.Config, tierEditName)
		if tier == nil {
			return fmt.Errorf("tier %q not found", tierEditName)
		}

		// Only update fields that were explicitly set
		if cmd.Flags().Changed("value") {
			tier.Value = tierEditValue
		}
		if cmd.Flags().Changed("currency") {
			tier.Currency = tierEditCurrency
		}
		if cmd.Flags().Changed("style") {
			// Validate the new style exists
			if core.FindStyleByName(&app.Config, tierEditStyle) == nil {
				return fmt.Errorf("style %q not found in style library", tierEditStyle)
			}
			tier.StyleName = tierEditStyle
		}

		app.UpdateConfig(app.Config)
		fmt.Printf("Updated tier %q (%.2f %s → style %q).\n",
			tier.Name, tier.Value, tier.Currency, tier.StyleName)
		return nil
	},
}

var tierRemoveName string

var tiersRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a value tier",
	RunE: func(cmd *cobra.Command, args []string) error {
		if tierRemoveName == "" {
			return fmt.Errorf("--name is required")
		}

		if err := core.RemoveTier(&app.Config, tierRemoveName); err != nil {
			return err
		}

		app.UpdateConfig(app.Config)
		fmt.Printf("Removed tier %q.\n", tierRemoveName)
		return nil
	},
}

func init() {
	tiersAddCmd.Flags().StringVar(&tierAddName, "name", "", "Name for the new tier (required)")
	tiersAddCmd.Flags().Float64Var(&tierAddValue, "value", 1.0, "Value threshold")
	tiersAddCmd.Flags().StringVar(&tierAddCurrency, "currency", "Chaos", "Currency type (Chaos, Exalted, Divine)")
	tiersAddCmd.Flags().StringVar(&tierAddStyle, "style", "", "Style name to apply (required)")

	tiersEditCmd.Flags().StringVar(&tierEditName, "name", "", "Name of the tier to edit (required)")
	tiersEditCmd.Flags().Float64Var(&tierEditValue, "value", 0, "New value threshold")
	tiersEditCmd.Flags().StringVar(&tierEditCurrency, "currency", "", "New currency type")
	tiersEditCmd.Flags().StringVar(&tierEditStyle, "style", "", "New style name")

	tiersRemoveCmd.Flags().StringVar(&tierRemoveName, "name", "", "Name of the tier to remove (required)")

	tiersCmd.AddCommand(tiersListCmd)
	tiersCmd.AddCommand(tiersAddCmd)
	tiersCmd.AddCommand(tiersEditCmd)
	tiersCmd.AddCommand(tiersRemoveCmd)
}
