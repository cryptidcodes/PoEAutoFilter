//go:build windows
// +build windows

// tier_editor.go — Windows-only tier editing dialog using lxn/walk.

package windows

import (
	"fmt"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// RunTierDialog opens a modal dialog for editing a single value tier.
// Returns (1, nil) if accepted, (0, nil) if cancelled.
func RunTierDialog(owner walk.Form, tier *core.Tier, styles []core.Style) (int, error) {
	var dlg *walk.Dialog
	var nameEdit *walk.LineEdit
	var valueEdit *walk.LineEdit
	var currencyBox *walk.ComboBox
	var styleBox *walk.ComboBox
	var acceptBtn, cancelBtn *walk.PushButton

	styleNames := make([]string, len(styles))
	for i, s := range styles {
		styleNames[i] = s.Name
	}

	currencies := []string{"Chaos", "Exalted", "Divine", "Mirror"}

	return Dialog{
		AssignTo:      &dlg,
		Title:         "Edit Value Tier",
		DefaultButton: &acceptBtn,
		CancelButton:  &cancelBtn,
		MinSize:       Size{Width: 350, Height: 250},
		Layout:        VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Tier Name:"},
					LineEdit{AssignTo: &nameEdit, Text: tier.Name},
					Label{Text: "Value threshold:"},
					LineEdit{AssignTo: &valueEdit, Text: fmt.Sprintf("%.2f", tier.Value)},
					Label{Text: "Currency Type:"},
					ComboBox{AssignTo: &currencyBox, Model: currencies, Value: tier.Currency},
					Label{Text: "Style to Apply:"},
					ComboBox{AssignTo: &styleBox, Model: styleNames, Value: tier.StyleName},
				},
			},
			VSpacer{},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{AssignTo: &acceptBtn, Text: "OK", OnClicked: func() {
						tier.Name = nameEdit.Text()
						var val float64
						if n, err := fmt.Sscanf(valueEdit.Text(), "%f", &val); err == nil && n > 0 {
							tier.Value = val
						}
						tier.Currency = currencyBox.Text()
						tier.StyleName = styleBox.Text()
						dlg.Accept()
					}},
					PushButton{AssignTo: &cancelBtn, Text: "Cancel", OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}.Run(owner)
}
