package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ShowGUI initializes and runs the Fyne GUI application.
func ShowGUI() {
	a := app.New()
	w := a.NewWindow("PoEAutoFilter Config")
	w.Resize(fyne.NewSize(800, 600))

	cfg, err := LoadConfig()
	if err != nil {
		cfg = Config{League: "Standard"}
	}

	// Shared Log Area
	logEntry := widget.NewMultiLineEntry()
	logEntry.Disable()
	logEntry.SetMinRowsVisible(4) // Half as tall as before
	logFunc := func(msg string) {
		logEntry.SetText(logEntry.Text + msg)
		logEntry.Refresh()
		// Auto-scroll to bottom
		logEntry.CursorRow = len(strings.Split(logEntry.Text, "\n"))
		logEntry.Refresh()
	}

	// 1. General Settings Tab
	leagueEntry := widget.NewEntry()
	leagueEntry.SetText(cfg.League)

	baseFilePathEntry := widget.NewEntry()
	baseFilePathEntry.SetText(cfg.BaseFilePath)
	baseBrowseButton := widget.NewButton("Browse...", func() {
		d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader != nil {
				baseFilePathEntry.SetText(reader.URI().Path())
				cfg.BaseFilePath = reader.URI().Path()
			}
		}, w)
		d.Resize(fyne.NewSize(800, 600))
		d.Show()
	})

	filePathEntry := widget.NewEntry()
	filePathEntry.SetText(cfg.FilePath)
	browseButton := widget.NewButton("Browse...", func() {
		d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader != nil {
				filePathEntry.SetText(reader.URI().Path())
				cfg.FilePath = reader.URI().Path()
			}
		}, w)
		d.Resize(fyne.NewSize(800, 600))
		d.Show()
	})

	customRulesEntry := widget.NewMultiLineEntry()
	customRulesEntry.SetText(cfg.Override)
	customRulesEntry.SetMinRowsVisible(12) // Twice as tall as before (was 4, plus more)

	settingsFields := container.NewVBox(
		widget.NewLabel("League:"), leagueEntry,
		widget.NewLabel("Base Filter File (Template):"),
		container.NewBorder(nil, nil, nil, baseBrowseButton, baseFilePathEntry),
		widget.NewLabel("Output Filter File (Target):"),
		container.NewBorder(nil, nil, nil, browseButton, filePathEntry),
		widget.NewLabel("Custom Rules Override:"), customRulesEntry,
	)

	startButton := widget.NewButton("Save & Start AutoFilter", func() {
		cfg.League = leagueEntry.Text
		cfg.Override = customRulesEntry.Text
		cfg.BaseFilePath = baseFilePathEntry.Text
		cfg.FilePath = filePathEntry.Text
		SaveConfig(cfg)

		leagueEntry.Disable()
		customRulesEntry.Disable()
		baseBrowseButton.Disable()
		browseButton.Disable()
		baseFilePathEntry.Disable()
		filePathEntry.Disable()

		go runBot(cfg, logFunc)
	})

	generalTabContent := container.NewBorder(
		container.NewVBox(settingsFields, startButton),
		nil, nil, nil,
		container.NewBorder(
			widget.NewLabelWithStyle("Activity Log:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			logEntry,
		),
	)

	// 2. Style Library Tab
	styleLibraryContainer := container.NewVBox()
	var refreshStyleLibrary func()
	var refreshTiers func()

	refreshStyleLibrary = func() {
		styleLibraryContainer.Objects = nil
		for i := range cfg.StyleLibrary {
			idx := i
			style := &cfg.StyleLibrary[idx]

			nameEntry := widget.NewEntry()
			nameEntry.SetText(style.Name)
			nameEntry.OnChanged = func(s string) {
				style.Name = s
				refreshTiers() // Sync name changes
			}

			previewLabel := widget.NewLabel(style.ToFilterLines())
			previewLabel.Hide()

			toggleButton := widget.NewButton("Preview Rules", func() {
				if previewLabel.Visible() {
					previewLabel.Hide()
				} else {
					previewLabel.Show()
				}
				styleLibraryContainer.Refresh()
			})

			editButton := widget.NewButton("Edit Style", func() {
				showStyleEditor(w, style, func() {
					previewLabel.SetText(style.ToFilterLines())
				})
			})

			removeButton := widget.NewButton("Delete", func() {
				cfg.StyleLibrary = append(cfg.StyleLibrary[:idx], cfg.StyleLibrary[idx+1:]...)
				refreshStyleLibrary()
				refreshTiers()
			})

			card := container.NewVBox(
				container.NewBorder(nil, nil, widget.NewLabel("Style Name:"), container.NewHBox(toggleButton, editButton, removeButton), nameEntry),
				previewLabel,
				widget.NewSeparator(),
			)
			styleLibraryContainer.Add(card)
		}
		styleLibraryContainer.Add(widget.NewButton("Add New Style", func() {
			cfg.StyleLibrary = append(cfg.StyleLibrary, Style{Name: "New Style"})
			refreshStyleLibrary()
			refreshTiers()
		}))
		styleLibraryContainer.Refresh()
	}

	// 3. Value Tiers Tab
	tiersContainer := container.NewVBox()
	refreshTiers = func() {
		tiersContainer.Objects = nil

		styleNames := []string{}
		for _, s := range cfg.StyleLibrary {
			styleNames = append(styleNames, s.Name)
		}

		for i := range cfg.Tiers {
			idx := i
			tier := &cfg.Tiers[idx]

			nameEntry := widget.NewEntry()
			nameEntry.SetText(tier.Name)
			nameEntry.OnChanged = func(s string) { tier.Name = s }

			valEntry := widget.NewEntry()
			valEntry.SetText(fmt.Sprintf("%.2f", tier.Value))
			valEntry.OnChanged = func(s string) {
				if v, err := strconv.ParseFloat(s, 64); err == nil {
					tier.Value = v
				}
			}

			currSelect := widget.NewSelect([]string{"Chaos", "Exalted", "Divine"}, func(s string) {
				tier.Currency = s
			})
			currSelect.SetSelected(tier.Currency)

			styleNames := []string{}
			for _, s := range cfg.StyleLibrary {
				styleNames = append(styleNames, s.Name)
			}
			styleSelect := widget.NewSelect(styleNames, func(s string) {
				tier.StyleName = s
			})
			styleSelect.SetSelected(tier.StyleName)

			removeButton := widget.NewButton("X", func() {
				cfg.Tiers = append(cfg.Tiers[:idx], cfg.Tiers[idx+1:]...)
				refreshTiers()
			})

			rightGroup := container.NewHBox(
				widget.NewLabel("Val:"),
				container.NewGridWrap(fyne.NewSize(100, 36), valEntry),
				currSelect,
				widget.NewLabel("Style:"), styleSelect,
				removeButton,
			)

			card := container.NewVBox(
				container.NewBorder(nil, nil, widget.NewLabel("Tier:"), rightGroup, nameEntry),
				widget.NewSeparator(),
			)
			tiersContainer.Add(card)
		}
		tiersContainer.Add(widget.NewButton("Add New Value Tier", func() {
			cfg.Tiers = append(cfg.Tiers, Tier{Name: "New Tier", Value: 1.0, Currency: "Chaos"})
			refreshTiers()
		}))
		tiersContainer.Refresh()
	}

	refreshStyleLibrary()
	refreshTiers()

	// Final Layout
	tabs := container.NewAppTabs(
		container.NewTabItem("General", generalTabContent),
		container.NewTabItem("Style Library", container.NewVScroll(styleLibraryContainer)),
		container.NewTabItem("Value Tiers", container.NewVScroll(tiersContainer)),
	)

	w.SetContent(tabs)
	w.ShowAndRun()
}

// showStyleEditor opens a modal to edit the actions of a style.
func showStyleEditor(w fyne.Window, style *Style, onComplete func()) {
	content := container.NewVBox()

	actionTypes := []string{"SetFontSize", "SetTextColor", "SetBorderColor", "SetBackgroundColor", "PlayAlertSound", "PlayEffect", "MinimapIcon"}

	// Constants for dropdowns
	iconSizes := []string{"Large", "Medium", "Small"}
	iconColours := []string{"Red", "Green", "Blue", "Brown", "White", "Yellow", "Cyan", "Grey", "Orange", "Pink", "Purple"}
	iconShapes := []string{"Circle", "Diamond", "Hexagon", "Square", "Star", "Triangle", "Cross", "Moon", "Raindrop", "Kite", "Pentagon", "UpsideDownHouse"}
	soundIDs := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16"}

	var refreshActions func()
	refreshActions = func() {
		content.Objects = nil
		for i := range style.Actions {
			idx := i
			action := &style.Actions[idx]

			typeLabel := widget.NewLabel(action.Type)
			typeLabel.TextStyle = fyne.TextStyle{Bold: true}

			// Dynamic Input based on action type
			var inputWidget fyne.CanvasObject
			switch action.Type {
			case "SetFontSize":
				val := 32.0
				if len(action.Values) > 0 {
					if v, err := strconv.ParseFloat(action.Values[0], 64); err == nil {
						val = v
					}
				}
				slider := widget.NewSlider(1, 45)
				slider.Value = val
				label := widget.NewLabel(fmt.Sprintf("%d", int(val)))
				slider.OnChanged = func(f float64) {
					action.Values = []string{fmt.Sprintf("%d", int(f))}
					label.SetText(fmt.Sprintf("%d", int(f)))
				}
				// Use Border to make the slider as wide as possible
				inputWidget = container.NewBorder(nil, nil, nil, label, slider)

			case "SetTextColor", "SetBorderColor", "SetBackgroundColor":
				// Color Picker implementation
				entry := widget.NewEntry()
				entry.SetText(strings.Join(action.Values, " "))
				entry.OnChanged = func(s string) { action.Values = strings.Fields(s) }

				colorBtn := widget.NewButton("Pick Color", func() {
					cp := dialog.NewColorPicker("Select Color", "Pick", func(c color.Color) {
						r, g, b, a := c.RGBA()
						// RGBA() returns 0-65535, scale to 0-255
						action.Values = []string{
							strconv.Itoa(int(r >> 8)),
							strconv.Itoa(int(g >> 8)),
							strconv.Itoa(int(b >> 8)),
							strconv.Itoa(int(a >> 8)),
						}
						entry.SetText(strings.Join(action.Values, " "))
					}, w)
					cp.Advanced = true
					cp.Show()
				})
				inputWidget = container.NewBorder(nil, nil, nil, colorBtn, entry)

			case "PlayAlertSound":
				// ID Dropdown and Volume Slider
				id := "1"
				if len(action.Values) > 0 {
					id = action.Values[0]
				}
				vol := 50.0
				if len(action.Values) > 1 {
					if v, err := strconv.ParseFloat(action.Values[1], 64); err == nil {
						vol = v
					}
				}

				idSelect := newCompactSelect(soundIDs, id, func(s string) {
					if len(action.Values) < 1 {
						action.Values = []string{s, "50"}
					} else {
						action.Values[0] = s
					}
				})

				volLabel := widget.NewLabel(fmt.Sprintf("Vol: %d", int(vol)))
				volSlider := widget.NewSlider(0, 300)
				volSlider.Value = vol
				volSlider.OnChanged = func(f float64) {
					if len(action.Values) < 2 {
						action.Values = append(action.Values, fmt.Sprintf("%d", int(f)))
					} else {
						action.Values[1] = fmt.Sprintf("%d", int(f))
					}
					volLabel.SetText(fmt.Sprintf("Vol: %d", int(f)))
				}
				// Widen vol slider
				idItem := container.NewHBox(widget.NewLabel("ID:"), idSelect, volLabel)
				inputWidget = container.NewBorder(nil, nil, idItem, nil, volSlider)

			case "PlayEffect":
				// Color Dropdown and Temp Toggle
				col := "White"
				if len(action.Values) > 0 {
					col = action.Values[0]
				}
				temp := ""
				if len(action.Values) > 1 {
					temp = action.Values[1]
				}

				colSelect := newCompactSelect(iconColours, col, func(s string) {
					if len(action.Values) < 1 {
						action.Values = []string{s}
					} else {
						action.Values[0] = s
					}
				})

				tempCheck := widget.NewCheck("Temp", func(b bool) {
					if b {
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
				tempCheck.SetChecked(temp == "Temp")
				inputWidget = container.NewHBox(colSelect, tempCheck)

			case "MinimapIcon":
				// Size, Color, Shape Dropdowns
				sz, cl, sh := "Large", "White", "Circle"
				if len(action.Values) > 0 {
					sz = action.Values[0]
				}
				if len(action.Values) > 1 {
					cl = action.Values[1]
				}
				if len(action.Values) > 2 {
					sh = action.Values[2]
				}

				szSelect := newCompactSelect(iconSizes, sz, func(s string) {
					ensureActionValues(action, 0)
					action.Values[0] = s
				})

				clSelect := newCompactSelect(iconColours, cl, func(s string) {
					ensureActionValues(action, 1)
					action.Values[1] = s
				})

				shSelect := newCompactSelect(iconShapes, sh, func(s string) {
					ensureActionValues(action, 2)
					action.Values[2] = s
				})
				inputWidget = container.NewHBox(szSelect, clSelect, shSelect)

			default:
				entry := widget.NewEntry()
				entry.SetText(strings.Join(action.Values, " "))
				entry.OnChanged = func(s string) { action.Values = strings.Fields(s) }
				inputWidget = entry
			}

			delButton := widget.NewButton("X", func() {
				style.Actions = append(style.Actions[:idx], style.Actions[idx+1:]...)
				refreshActions()
			})

			row := container.NewVBox(
				container.NewBorder(nil, nil, typeLabel, delButton, inputWidget),
				widget.NewSeparator(),
			)
			content.Add(row)
		}

		addSelect := widget.NewSelect(actionTypes, func(s string) {
			action := FilterAction{Type: s}
			switch s {
			case "SetFontSize":
				action.Values = []string{"32"}
			case "SetTextColor":
				action.Values = []string{"255", "255", "255", "255"}
			case "SetBorderColor", "SetBackgroundColor":
				action.Values = []string{"0", "0", "0", "255"}
			case "PlayAlertSound":
				action.Values = []string{"1", "50"}
			case "PlayEffect":
				action.Values = []string{"White"}
			case "MinimapIcon":
				action.Values = []string{"Large", "White", "Circle"}
			}
			style.Actions = append(style.Actions, action)
			refreshActions()
		})
		addSelect.PlaceHolder = "Add Action..."
		content.Add(addSelect)
		content.Refresh()
	}
	refreshActions()

	scroll := container.NewVScroll(content)

	d := dialog.NewCustomConfirm("Edit Style: "+style.Name, "OK", "Cancel", scroll, func(ok bool) {
		if ok {
			onComplete()
		}
	}, w)
	d.Resize(fyne.NewSize(650, 500))
	d.Show()
}

func ensureActionValues(action *FilterAction, minIndex int) {
	for len(action.Values) <= minIndex {
		action.Values = append(action.Values, "")
	}
}

// newCompactSelect creates a dropdown-like button that opens a scrollable list with limited height.
func newCompactSelect(options []string, selected string, onChanged func(string)) fyne.CanvasObject {
	btn := widget.NewButton(selected, nil)
	btn.Icon = theme.MenuIcon()
	// Default alignment/icon placement to be consistent with Fyne standards

	btn.OnTapped = func() {
		var pop *widget.PopUp
		list := widget.NewList(
			func() int { return len(options) },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				obj.(*widget.Label).SetText(options[id])
			},
		)
		list.OnSelected = func(id widget.ListItemID) {
			val := options[id]
			btn.SetText(val)
			onChanged(val)
			pop.Hide()
		}

		// Wrap list in a grid wrap to fix height (approx 5 items * 32px)
		content := container.NewGridWrap(fyne.NewSize(150, 160), list)

		pop = widget.NewPopUp(content, fyne.CurrentApp().Driver().CanvasForObject(btn))
		pop.ShowAtPosition(fyne.CurrentApp().Driver().AbsolutePositionForObject(btn))
	}

	return btn
}
