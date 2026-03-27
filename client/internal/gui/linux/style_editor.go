//go:build linux

package linux

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// RunStyleEditor opens a GTK dialog to edit a Style.
func RunStyleEditor(parent *gtk.Window, style *core.Style, onSave func()) {
	dlg := gtk.NewDialogWithFlags("Advanced Style Editor", parent, gtk.DialogModal|gtk.DialogDestroyWithParent)
	dlg.SetDefaultSize(700, 500)
	
	content := dlg.ContentArea()
	
	mainBox := gtk.NewBox(gtk.OrientationVertical, 10)
	mainBox.SetMarginTop(10)
	mainBox.SetMarginBottom(10)
	mainBox.SetMarginStart(10)
	mainBox.SetMarginEnd(10)

	// Header
	headerBox := gtk.NewBox(gtk.OrientationHorizontal, 5)
	nameLabel := gtk.NewLabel("Style Name:")
	nameEntry := gtk.NewEntry()
	nameEntry.SetText(style.Name)
	nameEntry.SetHExpand(true)
	headerBox.Append(nameLabel)
	headerBox.Append(nameEntry)
	mainBox.Append(headerBox)

	// Split view
	split := gtk.NewPaned(gtk.OrientationHorizontal)
	split.SetVExpand(true)

	// Left Pane (Actions List)
	leftBox := gtk.NewBox(gtk.OrientationVertical, 5)
	
	actionListWrapper := gtk.NewScrolledWindow()
	actionListWrapper.SetVExpand(true)
	
	actionList := gtk.NewListBox()
	actionList.SetSelectionMode(gtk.SelectionSingle)
	actionListWrapper.SetChild(actionList)

	// A helper to rebuild the list
	rebuildActionList := func(actions []core.FilterAction) {
		// Clean existing
		for child := actionList.FirstChild(); child != nil; child = actionList.FirstChild() {
			actionList.Remove(child)
		}
		for _, action := range actions {
			lbl := gtk.NewLabel(action.Type)
			lbl.SetXAlign(0)
			lbl.SetMarginStart(5)
			lbl.SetMarginTop(5)
			lbl.SetMarginBottom(5)
			actionList.Append(lbl)
		}
	}

	tempActions := make([]core.FilterAction, len(style.Actions))
	copy(tempActions, style.Actions)
	rebuildActionList(tempActions)

	leftBox.Append(gtk.NewLabel("Actions:"))
	leftBox.Append(actionListWrapper)

	// Add/Remove buttons
	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 5)
	
	actionTypeCombo := gtk.NewComboBoxText()
	types := []string{
		"SetFontSize", "SetTextColor", "SetBorderColor", "SetBackgroundColor",
		"PlayAlertSound", "PlayEffect", "MinimapIcon", "DisableDropSound", "Custom",
	}
	for _, t := range types {
		actionTypeCombo.Append(t, t)
	}
	actionTypeCombo.SetActiveID(types[0])
	actionTypeCombo.SetHExpand(true)
	
	addBtn := gtk.NewButtonWithLabel("Add")
	addBtn.SetHExpand(true)
	removeBtn := gtk.NewButtonWithLabel("Remove")
	removeBtn.SetHExpand(true)
	
	btnBox.Append(addBtn)
	btnBox.Append(removeBtn)
	
	leftBox.Append(actionTypeCombo)
	leftBox.Append(btnBox)

	split.SetStartChild(leftBox)

	// Right Pane (Editor)
	rightBox := gtk.NewBox(gtk.OrientationVertical, 5)
	rightBox.SetMarginStart(10)
	editorContainer := gtk.NewBox(gtk.OrientationVertical, 5)
	
	rightScroll := gtk.NewScrolledWindow()
	rightScroll.SetVExpand(true)
	rightScroll.SetHExpand(true)
	rightScroll.SetChild(editorContainer)
	rightBox.Append(rightScroll)

	split.SetEndChild(rightBox)
	split.SetPosition(250) // Default split pos

	mainBox.Append(split)
	content.Append(mainBox)

	// Handle selection change
	actionList.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		// Clear right pane
		for child := editorContainer.FirstChild(); child != nil; child = editorContainer.FirstChild() {
			editorContainer.Remove(child)
		}

		if row == nil {
			editorContainer.Append(gtk.NewLabel("Select an action to edit..."))
			return
		}
		
		idx := row.Index()
		if idx < 0 || idx >= len(tempActions) {
			return
		}
		
		a := &tempActions[idx]
		
		lbl := gtk.NewLabel(fmt.Sprintf("Editing: %s", a.Type))
		lbl.SetXAlign(0)
		editorContainer.Append(lbl)

		switch a.Type {
		case "SetFontSize":
			createFontSizeEditor(editorContainer, a)
		case "SetTextColor", "SetBorderColor", "SetBackgroundColor":
			createColorEditor(editorContainer, a)
		case "PlayAlertSound":
			createSoundEditor(editorContainer, a)
		case "PlayEffect":
			createEffectEditor(editorContainer, a)
		case "MinimapIcon":
			createMinimapIconEditor(editorContainer, a)
		default:
			createGenericEditor(editorContainer, a)
		}
	})

	addBtn.ConnectClicked(func() {
		newType := actionTypeCombo.ActiveText()
		newAction := core.FilterAction{Type: newType, Values: getDefaultValuesForType(newType)}
		tempActions = append(tempActions, newAction)
		rebuildActionList(tempActions)
		// Select last row
		row := actionList.RowAtIndex(len(tempActions) - 1)
		actionList.SelectRow(row)
	})

	removeBtn.ConnectClicked(func() {
		row := actionList.SelectedRow()
		if row != nil {
			idx := row.Index()
			if idx >= 0 && idx < len(tempActions) {
				tempActions = append(tempActions[:idx], tempActions[idx+1:]...)
				rebuildActionList(tempActions)
			}
		}
	})

	dlg.AddButton("Cancel", int(gtk.ResponseCancel))
	dlg.AddButton("Save Style", int(gtk.ResponseAccept))

	dlg.ConnectResponse(func(response int) {
		if response == int(gtk.ResponseAccept) {
			style.Name = nameEntry.Text()
			style.Actions = tempActions
			onSave()
		}
		dlg.Destroy()
	})

	dlg.Show()
	
	// Initial selection
	if len(tempActions) > 0 {
		actionList.SelectRow(actionList.RowAtIndex(0))
	} else {
		editorContainer.Append(gtk.NewLabel("Select an action to edit..."))
	}
}

func getDefaultValuesForType(t string) []string {
	switch t {
	case "SetFontSize":
		return []string{"32"}
	case "SetTextColor":
		return []string{"255", "255", "255", "255"}
	case "SetBorderColor":
		return []string{"255", "0", "0", "255"}
	case "SetBackgroundColor":
		return []string{"0", "0", "0", "255"}
	case "PlayAlertSound":
		return []string{"1", "300"}
	case "PlayEffect":
		return []string{"White"}
	case "MinimapIcon":
		return []string{"0", "White", "Circle"}
	case "DisableDropSound":
		return []string{}
	default:
		return []string{}
	}
}

// Editor Builders for Right Pane

func createFontSizeEditor(parent *gtk.Box, action *core.FilterAction) {
	val := 32
	if len(action.Values) > 0 {
		if v, err := strconv.Atoi(action.Values[0]); err == nil {
			val = v
		}
	} else {
		action.Values = []string{"32"}
	}

	hb := gtk.NewBox(gtk.OrientationHorizontal, 5)
	hb.Append(gtk.NewLabel("Size:"))

	adj := gtk.NewAdjustment(float64(val), 18, 50, 1, 5, 0)
	scale := gtk.NewScale(gtk.OrientationHorizontal, adj)
	scale.SetHExpand(true)
	scale.SetDrawValue(true)
	
	// Use ValueChanged signal on Adjustment instead
	adj.ConnectValueChanged(func() {
		action.Values[0] = fmt.Sprintf("%d", int(adj.Value()))
	})

	hb.Append(scale)
	parent.Append(hb)
}

func createColorEditor(parent *gtk.Box, action *core.FilterAction) {
	r, g, b, a := 255, 255, 255, 255
	if len(action.Values) > 0 { r, _ = strconv.Atoi(action.Values[0]) }
	if len(action.Values) > 1 { g, _ = strconv.Atoi(action.Values[1]) }
	if len(action.Values) > 2 { b, _ = strconv.Atoi(action.Values[2]) }
	if len(action.Values) > 3 { a, _ = strconv.Atoi(action.Values[3]) }
	action.Values = []string{strconv.Itoa(r), strconv.Itoa(g), strconv.Itoa(b), strconv.Itoa(a)}

	grid := gtk.NewGrid()
	grid.SetRowSpacing(5)
	grid.SetColumnSpacing(5)

	labels := []string{"Red (0-255):", "Green (0-255):", "Blue (0-255):", "Alpha (0-255):"}
	vals := []int{r, g, b, a}

	for i := 0; i < 4; i++ {
		lbl := gtk.NewLabel(labels[i])
		lbl.SetXAlign(0)
		entry := gtk.NewEntry()
		entry.SetText(strconv.Itoa(vals[i]))
		
		idx := i
		// Listen to changes
		entry.ConnectChanged(func() {
			action.Values[idx] = entry.Text()
		})

		grid.Attach(lbl, 0, i, 1, 1)
		grid.Attach(entry, 1, i, 1, 1)
	}
	parent.Append(grid)
}

func createSoundEditor(parent *gtk.Box, action *core.FilterAction) {
	id, vol := "1", "100"
	if len(action.Values) > 0 { id = action.Values[0] }
	if len(action.Values) > 1 { vol = action.Values[1] }
	if len(action.Values) < 2 { action.Values = []string{id, vol} }

	grid := gtk.NewGrid()
	grid.SetRowSpacing(5)
	grid.SetColumnSpacing(5)

	lblId := gtk.NewLabel("Sound ID:")
	lblId.SetXAlign(0)
	comboId := gtk.NewComboBoxText()
	ids := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "Shaper", "Elder"}
	for _, i := range ids {
		comboId.Append(i, i)
	}
	comboId.SetActiveID(id)
	grid.Attach(lblId, 0, 0, 1, 1)
	grid.Attach(comboId, 1, 0, 1, 1)

	lblVol := gtk.NewLabel("Volume (0-300):")
	lblVol.SetXAlign(0)
	entryVol := gtk.NewEntry()
	entryVol.SetText(vol)
	grid.Attach(lblVol, 0, 1, 1, 1)
	grid.Attach(entryVol, 1, 1, 1, 1)

	comboId.ConnectChanged(func() { action.Values[0] = comboId.ActiveText() })
	entryVol.ConnectChanged(func() { action.Values[1] = entryVol.Text() })

	parent.Append(grid)
}

func createEffectEditor(parent *gtk.Box, action *core.FilterAction) {
	color, temp := "White", ""
	if len(action.Values) > 0 { color = action.Values[0] }
	if len(action.Values) > 1 { temp = action.Values[1] }
	if len(action.Values) < 1 { action.Values = []string{color} }

	hb := gtk.NewBox(gtk.OrientationHorizontal, 5)
	hb.Append(gtk.NewLabel("Color:"))

	combo := gtk.NewComboBoxText()
	colors := []string{"Red", "Green", "Blue", "Brown", "White", "Yellow", "Cyan", "Grey", "Orange", "Pink", "Purple"}
	for _, c := range colors {
		combo.Append(c, c)
	}
	combo.SetActiveID(color)
	hb.Append(combo)

	check := gtk.NewCheckButtonWithLabel("Temporary")
	check.SetActive(temp == "Temp")
	hb.Append(check)

	combo.ConnectChanged(func() { action.Values[0] = combo.ActiveText() })
	check.ConnectToggled(func() {
		if check.Active() {
			if len(action.Values) < 2 {
				action.Values = append(action.Values, "Temp")
			} else {
				action.Values[1] = "Temp"
			}
		} else {
			if len(action.Values) > 1 {
				action.Values = action.Values[:1]
			}
		}
	})

	parent.Append(hb)
}

func createMinimapIconEditor(parent *gtk.Box, action *core.FilterAction) {
	size, color, shape := "0", "White", "Circle"
	if len(action.Values) > 0 { size = action.Values[0] }
	if len(action.Values) > 1 { color = action.Values[1] }
	if len(action.Values) > 2 { shape = action.Values[2] }
	action.Values = []string{size, color, shape}

	grid := gtk.NewGrid()
	grid.SetRowSpacing(5)
	grid.SetColumnSpacing(5)

	lblSize := gtk.NewLabel("Size:")
	lblSize.SetXAlign(0)
	comboSize := gtk.NewComboBoxText()
	sizes := []string{"0", "1", "2"}
	for _, s := range sizes {
		comboSize.Append(s, s)
	}
	comboSize.SetActiveID(size)
	grid.Attach(lblSize, 0, 0, 1, 1)
	grid.Attach(comboSize, 1, 0, 1, 1)

	lblColor := gtk.NewLabel("Color:")
	lblColor.SetXAlign(0)
	comboColor := gtk.NewComboBoxText()
	colors := []string{"Red", "Green", "Blue", "Brown", "White", "Yellow", "Cyan", "Grey", "Orange", "Pink", "Purple"}
	for _, c := range colors {
		comboColor.Append(c, c)
	}
	comboColor.SetActiveID(color)
	grid.Attach(lblColor, 0, 1, 1, 1)
	grid.Attach(comboColor, 1, 1, 1, 1)

	lblShape := gtk.NewLabel("Shape:")
	lblShape.SetXAlign(0)
	comboShape := gtk.NewComboBoxText()
	shapes := []string{"Circle", "Diamond", "Hexagon", "Square", "Star", "Triangle", "Cross", "Moon", "Raindrop", "Kite", "Pentagon", "UpsideDownHouse"}
	for _, s := range shapes {
		comboShape.Append(s, s)
	}
	comboShape.SetActiveID(shape)
	grid.Attach(lblShape, 0, 2, 1, 1)
	grid.Attach(comboShape, 1, 2, 1, 1)

	comboSize.ConnectChanged(func() { action.Values[0] = comboSize.ActiveText() })
	comboColor.ConnectChanged(func() { action.Values[1] = comboColor.ActiveText() })
	comboShape.ConnectChanged(func() { action.Values[2] = comboShape.ActiveText() })

	parent.Append(grid)
}

func createGenericEditor(parent *gtk.Box, action *core.FilterAction) {
	valStr := strings.Join(action.Values, " ")

	hb := gtk.NewBox(gtk.OrientationHorizontal, 5)
	hb.Append(gtk.NewLabel("Values (space separated):"))
	
	entry := gtk.NewEntry()
	entry.SetText(valStr)
	entry.SetHExpand(true)
	hb.Append(entry)

	entry.ConnectChanged(func() {
		action.Values = strings.Fields(entry.Text())
	})

	parent.Append(hb)
}
