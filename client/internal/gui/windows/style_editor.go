//go:build windows
// +build windows

// style_editor.go — Windows-only advanced style editor dialog using lxn/walk.

package windows

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// RunStyleEditor opens the advanced style editor dialog.
func RunStyleEditor(owner walk.Form, style *core.Style) (int, error) {
	defer LogPanic("RunStyleEditor", owner)

	var dlg *walk.Dialog
	var nameEdit *walk.LineEdit
	var actionList *walk.ListBox
	var editorContainer *walk.Composite
	var addBtn, removeBtn, acceptBtn, cancelBtn *walk.PushButton
	var actionTypeCombo *walk.ComboBox

	actionModel := NewActionListModel(style.Actions)

	actionTypes := []string{
		"SetFontSize", "SetTextColor", "SetBorderColor", "SetBackgroundColor",
		"PlayAlertSound", "PlayEffect", "MinimapIcon", "DisableDropSound", "Custom",
	}

	return Dialog{
		AssignTo:      &dlg,
		Title:         "Advanced Style Editor",
		DefaultButton: &acceptBtn,
		CancelButton:  &cancelBtn,
		MinSize:       Size{Width: 700, Height: 500},
		Layout:        VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "Style Name:"},
					LineEdit{AssignTo: &nameEdit, Text: style.Name},
				},
			},
			HSplitter{
				Children: []Widget{
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{Text: "Actions:"},
							ListBox{
								AssignTo: &actionList,
								Model:    actionModel,
								OnCurrentIndexChanged: func() {
									updateEditorPane(editorContainer, actionModel, actionList.CurrentIndex())
								},
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									ComboBox{
										AssignTo: &actionTypeCombo,
										Model:    actionTypes,
										Value:    actionTypes[0],
									},
									PushButton{AssignTo: &addBtn, Text: "Add", OnClicked: func() {
										newType := actionTypeCombo.Text()
										newAction := core.FilterAction{Type: newType, Values: getDefaultValuesForType(newType)}
										actionModel.Actions = append(actionModel.Actions, newAction)
										actionModel.PublishItemsReset()
										actionList.SetCurrentIndex(len(actionModel.Actions) - 1)
									}},
									PushButton{AssignTo: &removeBtn, Text: "Remove", OnClicked: func() {
										idx := actionList.CurrentIndex()
										if idx >= 0 && idx < len(actionModel.Actions) {
											actionModel.Actions = append(actionModel.Actions[:idx], actionModel.Actions[idx+1:]...)
											actionModel.PublishItemsReset()
											updateEditorPane(editorContainer, actionModel, -1)
										}
									}},
								},
							},
						},
					},
					Composite{
						AssignTo: &editorContainer,
						Layout:   VBox{},
						Children: []Widget{
							Label{Text: "Select an action to edit..."},
						},
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{AssignTo: &acceptBtn, Text: "Save Style", OnClicked: func() {
						style.Name = nameEdit.Text()
						style.Actions = actionModel.Actions
						dlg.Accept()
					}},
					PushButton{AssignTo: &cancelBtn, Text: "Cancel", OnClicked: func() {
						dlg.Cancel()
					}},
				},
			},
		},
	}.Run(owner)
}

func updateEditorPane(container *walk.Composite, model *ActionListModel, index int) {
	children := container.Children()
	if children.Len() > 0 {
		container.SetSuspended(true)
		defer container.SetSuspended(false)
		for i := children.Len() - 1; i >= 0; i-- {
			children.At(i).Dispose()
		}
	}

	if index < 0 || index >= len(model.Actions) {
		Label{Text: "Select an action to edit..."}.Create(NewBuilder(container))
		return
	}

	action := &model.Actions[index]
	builder := NewBuilder(container)
	Composite{
		Layout: VBox{},
		Children: []Widget{
			Label{Text: fmt.Sprintf("Editing: %s", action.Type), Font: Font{Bold: true, PointSize: 10}},
			VSpacer{Size: 5},
		},
	}.Create(builder)

	switch action.Type {
	case "SetFontSize":
		createFontSizeEditor(container, action)
	case "SetTextColor", "SetBorderColor", "SetBackgroundColor":
		createColorEditor(container, action)
	case "PlayAlertSound":
		createSoundEditor(container, action)
	case "PlayEffect":
		createEffectEditor(container, action)
	case "MinimapIcon":
		createMinimapIconEditor(container, action)
	default:
		createGenericEditor(container, action)
	}
}

func createFontSizeEditor(parent *walk.Composite, action *core.FilterAction) {
	val := 32
	if len(action.Values) > 0 {
		if v, err := strconv.Atoi(action.Values[0]); err == nil {
			val = v
		}
	} else {
		action.Values = []string{"32"}
	}

	cmp, err := walk.NewComposite(parent)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			cmp.Dispose()
		}
	}()

	layout := walk.NewHBoxLayout()
	cmp.SetLayout(layout)

	lblSize, _ := walk.NewLabel(cmp)
	lblSize.SetText("Size:")

	slider, _ := walk.NewSlider(cmp)
	slider.SetRange(18, 45)
	slider.SetValue(val)

	lblVal, _ := walk.NewLabel(cmp)
	lblVal.SetText(fmt.Sprintf("%d", val))
	lblVal.SetMinMaxSize(walk.Size{Width: 30, Height: 0}, walk.Size{Width: 30, Height: 0})

	slider.ValueChanged().Attach(func() {
		v := slider.Value()
		lblVal.SetText(fmt.Sprintf("%d", v))
		action.Values[0] = fmt.Sprintf("%d", v)
	})
}

func createColorEditor(parent *walk.Composite, action *core.FilterAction) {
	r, g, b, a := parseColor(action.Values)
	if len(action.Values) < 4 {
		action.Values = []string{strconv.Itoa(r), strconv.Itoa(g), strconv.Itoa(b), strconv.Itoa(a)}
	}

	builder := NewBuilder(parent)
	var rEdit, gEdit, bEdit, aEdit *walk.LineEdit

	Composite{
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "Red (0-255):"}, LineEdit{AssignTo: &rEdit, Text: strconv.Itoa(r), OnTextChanged: func() { updateColorVal(action, 0, rEdit.Text()) }},
			Label{Text: "Green (0-255):"}, LineEdit{AssignTo: &gEdit, Text: strconv.Itoa(g), OnTextChanged: func() { updateColorVal(action, 1, gEdit.Text()) }},
			Label{Text: "Blue (0-255):"}, LineEdit{AssignTo: &bEdit, Text: strconv.Itoa(b), OnTextChanged: func() { updateColorVal(action, 2, bEdit.Text()) }},
			Label{Text: "Alpha (0-255):"}, LineEdit{AssignTo: &aEdit, Text: strconv.Itoa(a), OnTextChanged: func() { updateColorVal(action, 3, aEdit.Text()) }},
		},
	}.Create(builder)
}

func updateColorVal(action *core.FilterAction, idx int, val string) {
	if _, err := strconv.Atoi(val); err == nil {
		if len(action.Values) > idx {
			action.Values[idx] = val
		}
	}
}

func createSoundEditor(parent *walk.Composite, action *core.FilterAction) {
	id, vol := "1", "100"
	if len(action.Values) > 0 {
		id = action.Values[0]
	}
	if len(action.Values) > 1 {
		vol = action.Values[1]
	}
	if len(action.Values) < 2 {
		action.Values = []string{id, vol}
	}

	builder := NewBuilder(parent)
	var idCombo *walk.ComboBox
	var volEdit *walk.LineEdit
	ids := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "Shaper", "Elder"}

	Composite{
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "Sound ID:"},
			ComboBox{AssignTo: &idCombo, Model: ids, Value: id, Editable: true, OnTextChanged: func() { action.Values[0] = idCombo.Text() }},
			Label{Text: "Volume (0-300):"},
			LineEdit{AssignTo: &volEdit, Text: vol, OnTextChanged: func() { action.Values[1] = volEdit.Text() }},
		},
	}.Create(builder)
}

func createEffectEditor(parent *walk.Composite, action *core.FilterAction) {
	color, temp := "White", ""
	if len(action.Values) > 0 {
		color = action.Values[0]
	}
	if len(action.Values) > 1 {
		temp = action.Values[1]
	}
	if len(action.Values) < 1 {
		action.Values = []string{color}
	}

	colors := []string{"Red", "Green", "Blue", "Brown", "White", "Yellow", "Cyan", "Grey", "Orange", "Pink", "Purple"}
	builder := NewBuilder(parent)
	var colorCombo *walk.ComboBox
	var tempCheck *walk.CheckBox

	Composite{
		Layout: HBox{},
		Children: []Widget{
			Label{Text: "Color:"},
			ComboBox{AssignTo: &colorCombo, Model: colors, Value: color, OnCurrentIndexChanged: func() { action.Values[0] = colorCombo.Text() }},
			CheckBox{AssignTo: &tempCheck, Text: "Temporary", Checked: temp == "Temp", OnCheckedChanged: func() {
				if tempCheck.Checked() {
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
			}},
		},
	}.Create(builder)
}

func createMinimapIconEditor(parent *walk.Composite, action *core.FilterAction) {
	size, color, shape := "0", "White", "Circle"
	if len(action.Values) > 0 {
		size = action.Values[0]
	}
	if len(action.Values) > 1 {
		color = action.Values[1]
	}
	if len(action.Values) > 2 {
		shape = action.Values[2]
	}
	action.Values = []string{size, color, shape}

	colors := []string{"Red", "Green", "Blue", "Brown", "White", "Yellow", "Cyan", "Grey", "Orange", "Pink", "Purple"}
	shapes := []string{"Circle", "Diamond", "Hexagon", "Square", "Star", "Triangle", "Cross", "Moon", "Raindrop", "Kite", "Pentagon", "UpsideDownHouse"}
	sizes := []string{"0 (Large)", "1 (Medium)", "2 (Small)"}

	builder := NewBuilder(parent)
	var sizeCombo, colorCombo, shapeCombo *walk.ComboBox

	Composite{
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "Size:"},
			ComboBox{AssignTo: &sizeCombo, Model: sizes, CurrentIndex: getIndex(sizes, size, 0), OnCurrentIndexChanged: func() {
				parts := strings.Fields(sizeCombo.Text())
				if len(parts) > 0 {
					action.Values[0] = parts[0]
				}
			}},
			Label{Text: "Color:"},
			ComboBox{AssignTo: &colorCombo, Model: colors, Value: color, OnCurrentIndexChanged: func() { action.Values[1] = colorCombo.Text() }},
			Label{Text: "Shape:"},
			ComboBox{AssignTo: &shapeCombo, Model: shapes, Value: shape, OnCurrentIndexChanged: func() { action.Values[2] = shapeCombo.Text() }},
		},
	}.Create(builder)
}

func createGenericEditor(parent *walk.Composite, action *core.FilterAction) {
	valStr := strings.Join(action.Values, " ")
	builder := NewBuilder(parent)
	var edit *walk.LineEdit

	Composite{
		Layout: HBox{},
		Children: []Widget{
			Label{Text: "Values (space separated):"},
			LineEdit{AssignTo: &edit, Text: valStr, OnTextChanged: func() { action.Values = strings.Fields(edit.Text()) }},
		},
	}.Create(builder)
}

// --- Helper types and functions ---

type ActionListModel struct {
	walk.ListModelBase
	Actions []core.FilterAction
}

func NewActionListModel(actions []core.FilterAction) *ActionListModel {
	return &ActionListModel{Actions: actions}
}

func (m *ActionListModel) ItemCount() int { return len(m.Actions) }

func (m *ActionListModel) Value(index int) interface{} {
	if index < 0 || index >= len(m.Actions) {
		return ""
	}
	return m.Actions[index].Type
}

func (m *ActionListModel) PublishItemsReset() {
	m.ListModelBase.PublishItemsReset()
}

func getDefaultValuesForType(t string) []string {
	switch t {
	case "SetFontSize":
		return []string{"32"}
	case "SetTextColor":
		return []string{"255", "255", "255", "255"}
	case "SetBorderColor":
		return []string{"0", "0", "0", "255"}
	case "SetBackgroundColor":
		return []string{"0", "0", "0", "255"}
	case "PlayAlertSound":
		return []string{"1", "100"}
	case "PlayEffect":
		return []string{"White", "Temp"}
	case "MinimapIcon":
		return []string{"0", "White", "Circle"}
	case "DisableDropSound":
		return []string{}
	default:
		return []string{}
	}
}

func parseColor(vals []string) (int, int, int, int) {
	r, g, b, a := 255, 255, 255, 255
	if len(vals) > 0 {
		r, _ = strconv.Atoi(vals[0])
	}
	if len(vals) > 1 {
		g, _ = strconv.Atoi(vals[1])
	}
	if len(vals) > 2 {
		b, _ = strconv.Atoi(vals[2])
	}
	if len(vals) > 3 {
		a, _ = strconv.Atoi(vals[3])
	}
	return r, g, b, a
}

func getIndex(arr []string, val string, def int) int {
	for i, v := range arr {
		if strings.HasPrefix(v, val) {
			return i
		}
	}
	return def
}
