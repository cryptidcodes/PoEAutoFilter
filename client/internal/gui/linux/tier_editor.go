//go:build linux

package linux

import (
	"fmt"
	"strconv"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// RunTierDialog opens a GTK dialog to edit a Tier.
func RunTierDialog(parent *gtk.Window, tier *core.Tier, styles []core.Style, onSave func()) {
	dlg := gtk.NewDialogWithFlags("Edit Value Tier", parent, gtk.DialogModal|gtk.DialogDestroyWithParent)
	
	content := dlg.ContentArea()
	grid := gtk.NewGrid()
	grid.SetMarginTop(10)
	grid.SetMarginBottom(10)
	grid.SetMarginStart(10)
	grid.SetMarginEnd(10)
	grid.SetRowSpacing(10)
	grid.SetColumnSpacing(10)

	// Name
	nameLabel := gtk.NewLabel("Tier Name:")
	nameLabel.SetHAlign(gtk.AlignStart)
	nameEntry := gtk.NewEntry()
	nameEntry.SetText(tier.Name)
	grid.Attach(nameLabel, 0, 0, 1, 1)
	grid.Attach(nameEntry, 1, 0, 1, 1)

	// Value
	valLabel := gtk.NewLabel("Value threshold:")
	valLabel.SetHAlign(gtk.AlignStart)
	valEntry := gtk.NewEntry()
	valEntry.SetText(fmt.Sprintf("%.2f", tier.Value))
	grid.Attach(valLabel, 0, 1, 1, 1)
	grid.Attach(valEntry, 1, 1, 1, 1)

	// Currency
	currLabel := gtk.NewLabel("Currency Type:")
	currLabel.SetHAlign(gtk.AlignStart)
	currCombo := gtk.NewComboBoxText()
	currencies := []string{"Chaos", "Exalted", "Divine", "Mirror"}
	for _, c := range currencies {
		currCombo.Append(c, c)
	}
	currCombo.SetActiveID(tier.Currency)
	grid.Attach(currLabel, 0, 2, 1, 1)
	grid.Attach(currCombo, 1, 2, 1, 1)

	// Style
	styleLabel := gtk.NewLabel("Style to Apply:")
	styleLabel.SetHAlign(gtk.AlignStart)
	styleCombo := gtk.NewComboBoxText()
	for _, s := range styles {
		styleCombo.Append(s.Name, s.Name)
	}
	styleCombo.SetActiveID(tier.StyleName)
	grid.Attach(styleLabel, 0, 3, 1, 1)
	grid.Attach(styleCombo, 1, 3, 1, 1)

	content.Append(grid)

	dlg.AddButton("Cancel", int(gtk.ResponseCancel))
	dlg.AddButton("OK", int(gtk.ResponseAccept))

	dlg.ConnectResponse(func(response int) {
		if response == int(gtk.ResponseAccept) {
			tier.Name = nameEntry.Text()
			if val, err := strconv.ParseFloat(valEntry.Text(), 64); err == nil {
				tier.Value = val
			}
			tier.Currency = currCombo.ActiveText()
			tier.StyleName = styleCombo.ActiveText()
			onSave()
		}
		dlg.Destroy()
	})

	dlg.Show()
}
