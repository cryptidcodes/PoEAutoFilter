//go:build windows
// +build windows

// Package windows provides the Windows native GUI for PoEAutoFilter using lxn/walk.
package windows

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// RunGUI is the Windows GUI entry point.
func RunGUI(app *core.App) {
	defer LogPanic("RunGUI", nil)

	var mw *walk.MainWindow
	var logText *walk.TextEdit
	var startBtn *walk.PushButton
	var styleTable *walk.TableView
	var tierTable *walk.TableView
	var basePathLE *walk.LineEdit
	var outputPathLE *walk.LineEdit
	var leagueLE *walk.LineEdit
	var overrideTE *walk.TextEdit

	// Table Models
	styleModel := NewStyleModel(app.Config.StyleLibrary)
	tierModel := NewTierModel(app.Config.Tiers)

	// Log Function Binding — routes app.Log() calls to the GUI Activity Log
	app.LogFunc = func(msg string) {
		if logText != nil {
			logText.Synchronize(func() {
				logText.AppendText(strings.ReplaceAll(msg, "\n", "\r\n"))
			})
		} else {
			fmt.Print(msg)
		}
	}

	if err := (MainWindow{
		AssignTo: &mw,
		Title:    "PoEAutoFilter Config",
		MinSize:  Size{Width: 800, Height: 600},
		Layout:   VBox{},
		Children: []Widget{
			TabWidget{
				Pages: []TabPage{
					// TAB 1: GENERAL SETTINGS
					{
						Title:  "General",
						Layout: VBox{},
						Children: []Widget{
							Composite{
								Layout: Grid{Columns: 3},
								Children: []Widget{
									Label{Text: "League:"},
									LineEdit{
										AssignTo: &leagueLE,
										Text:     app.Config.League,
										OnTextChanged: func() {
											if leagueLE != nil {
												app.Config.League = strings.TrimSpace(leagueLE.Text())
												app.UpdateConfig(app.Config)
											}
										},
									},
									HSpacer{},

									Label{Text: "Base Filter:"},
									LineEdit{
										AssignTo: &basePathLE,
										Text:     app.Config.BaseFilePath,
										ReadOnly: true,
									},
									PushButton{Text: "Browse...", OnClicked: func() {
										if mw == nil {
											return
										}
										cleanPath := filepath.FromSlash(filepath.Clean(app.Config.BaseFilePath))
										dlg := new(walk.FileDialog)
										dlg.Title = "Select Base Filter File"
										dlg.InitialDirPath = filepath.Dir(cleanPath)
										dlg.FilePath = cleanPath
										dlg.Filter = "Filter Files (*.filter)|*.filter|Text Files (*.txt)|*.txt|All Files (*.*)|*.*"

										ok, err := dlg.ShowOpen(mw)
										if err != nil {
											app.Log(fmt.Sprintf("Dialog Error: %v\n", err))
										}
										if ok {
											app.Config.BaseFilePath = filepath.ToSlash(dlg.FilePath)
											if basePathLE != nil {
												basePathLE.SetText(app.Config.BaseFilePath)
											}
											app.UpdateConfig(app.Config)
										}
									}},

									Label{Text: "Output Filter:"},
									LineEdit{
										AssignTo: &outputPathLE,
										Text:     app.Config.FilePath,
										ReadOnly: true,
									},
									PushButton{Text: "Browse...", OnClicked: func() {
										if mw == nil {
											return
										}
										cleanPath := filepath.FromSlash(filepath.Clean(app.Config.FilePath))
										dlg := new(walk.FileDialog)
										dlg.Title = "Select Output Filter File"
										dlg.InitialDirPath = filepath.Dir(cleanPath)
										dlg.FilePath = cleanPath
										dlg.Filter = "Filter Files (*.filter)|*.filter|Text Files (*.txt)|*.txt|All Files (*.*)|*.*"

										ok, err := dlg.ShowSave(mw)
										if err != nil {
											app.Log(fmt.Sprintf("Dialog Error: %v\n", err))
										}
										if ok {
											app.Config.FilePath = filepath.ToSlash(dlg.FilePath)
											if outputPathLE != nil {
												outputPathLE.SetText(app.Config.FilePath)
											}
											app.UpdateConfig(app.Config)
										}
									}},
								},
							},
							VSpacer{Size: 10},
							Label{Text: "Custom Rules Override:"},
							TextEdit{
								AssignTo: &overrideTE,
								Text:     app.Config.Override,
								MinSize:  Size{Height: 100},
								OnTextChanged: func() {
									if overrideTE != nil {
										app.Config.Override = overrideTE.Text()
										app.UpdateConfig(app.Config)
									}
								},
							},
							VSpacer{},
							PushButton{
								AssignTo: &startBtn,
								Text:     "Start AutoFilter",
								OnClicked: func() {
									if app.State.IsRunning {
										app.StopBot()
										startBtn.SetText("Start AutoFilter")
									} else {
										app.StartBot()
										startBtn.SetText("Stop AutoFilter")
									}
								},
							},
						},
					},

					// TAB 2: STYLE LIBRARY
					{
						Title:  "Style Library",
						Layout: VBox{},
						Children: []Widget{
							Composite{
								Layout: HBox{},
								Children: []Widget{
									PushButton{Text: "Add Style", OnClicked: func() {
										newStyle := core.Style{Name: "New Style", Actions: []core.FilterAction{}}
										app.Config.StyleLibrary = append(app.Config.StyleLibrary, newStyle)
										app.UpdateConfig(app.Config)

										// Refresh style model from updated config
										styleModel.Items = make([]*StyleItem, len(app.Config.StyleLibrary))
										for i := range app.Config.StyleLibrary {
											s := &app.Config.StyleLibrary[i]
											styleModel.Items[i] = &StyleItem{
												Name:    s.Name,
												Preview: s.ToFilterLines(),
												Style:   s,
											}
										}
										styleModel.PublishRowsReset()
									}},
									PushButton{Text: "Edit Selected", OnClicked: func() {
										idx := styleTable.CurrentIndex()
										if idx >= 0 && idx < len(styleModel.Items) {
											item := styleModel.Items[idx]
											if cmd, _ := RunStyleEditor(mw, item.Style); cmd == 1 {
												app.UpdateConfig(app.Config)
												item.Name = item.Style.Name
												item.Preview = item.Style.ToFilterLines()
												styleModel.PublishRowsReset()
											}
										}
									}},
									PushButton{Text: "Delete Style", OnClicked: func() {
										idx := styleTable.CurrentIndex()
										if idx >= 0 && idx < len(app.Config.StyleLibrary) {
											app.Config.StyleLibrary = append(app.Config.StyleLibrary[:idx], app.Config.StyleLibrary[idx+1:]...)
											styleModel.Items = append(styleModel.Items[:idx], styleModel.Items[idx+1:]...)
											styleModel.PublishRowsReset()
											app.UpdateConfig(app.Config)
										}
									}},
								},
							},
							TableView{
								AssignTo: &styleTable,
								Columns: []TableViewColumn{
									{Title: "Name", Width: 150},
									{Title: "Preview", Width: 300},
								},
								Model: styleModel,
								OnItemActivated: func() {
									idx := styleTable.CurrentIndex()
									if idx >= 0 {
										item := styleModel.Items[idx]
										if cmd, _ := RunStyleEditor(mw, item.Style); cmd == 1 {
											item.Name = item.Style.Name
											item.Preview = item.Style.ToFilterLines()
											styleModel.PublishRowsReset()
											app.UpdateConfig(app.Config)
										}
									}
								},
							},
						},
					},

					// TAB 3: VALUE TIERS
					{
						Title:  "Value Tiers",
						Layout: VBox{},
						Children: []Widget{
							Composite{
								Layout: HBox{},
								Children: []Widget{
									PushButton{Text: "Add Tier", OnClicked: func() {
										defaultStyle := "Default"
										if len(app.Config.StyleLibrary) > 0 {
											defaultStyle = app.Config.StyleLibrary[0].Name
										}
										newTier := core.Tier{Name: "New Tier", Value: 1.0, Currency: "Chaos", StyleName: defaultStyle}
										app.Config.Tiers = append(app.Config.Tiers, newTier)
										app.UpdateConfig(app.Config)

										// Refresh tier model from updated config
										tierModel.Items = make([]*TierItem, len(app.Config.Tiers))
										for i := range app.Config.Tiers {
											t := &app.Config.Tiers[i]
											tierModel.Items[i] = &TierItem{
												Name:     t.Name,
												Value:    fmt.Sprintf("%.2f", t.Value),
												Currency: t.Currency,
												Style:    t.StyleName,
												Tier:     t,
											}
										}
										tierModel.PublishRowsReset()
									}},
									PushButton{Text: "Edit Selected", OnClicked: func() {
										idx := tierTable.CurrentIndex()
										if idx >= 0 && idx < len(tierModel.Items) {
											item := tierModel.Items[idx]
											if cmd, _ := RunTierDialog(mw, item.Tier, app.Config.StyleLibrary); cmd == 1 {
												item.Name = item.Tier.Name
												item.Value = fmt.Sprintf("%.2f", item.Tier.Value)
												item.Currency = item.Tier.Currency
												item.Style = item.Tier.StyleName
												tierModel.PublishRowsReset()
												app.UpdateConfig(app.Config)
											}
										}
									}},
									PushButton{Text: "Delete Tier", OnClicked: func() {
										idx := tierTable.CurrentIndex()
										if idx >= 0 && idx < len(app.Config.Tiers) {
											app.Config.Tiers = append(app.Config.Tiers[:idx], app.Config.Tiers[idx+1:]...)
											tierModel.Items = append(tierModel.Items[:idx], tierModel.Items[idx+1:]...)
											tierModel.PublishRowsReset()
											app.UpdateConfig(app.Config)
										}
									}},
								},
							},
							TableView{
								AssignTo: &tierTable,
								Columns: []TableViewColumn{
									{Title: "Name", Width: 150},
									{Title: "Value", Width: 60},
									{Title: "Currency", Width: 80},
									{Title: "Style", Width: 150},
								},
								Model: tierModel,
								OnItemActivated: func() {
									idx := tierTable.CurrentIndex()
									if idx >= 0 {
										item := tierModel.Items[idx]
										if cmd, _ := RunTierDialog(mw, item.Tier, app.Config.StyleLibrary); cmd == 1 {
											item.Name = item.Tier.Name
											item.Value = fmt.Sprintf("%.2f", item.Tier.Value)
											item.Currency = item.Tier.Currency
											item.Style = item.Tier.StyleName
											tierModel.PublishRowsReset()
											app.UpdateConfig(app.Config)
										}
									}
								},
							},
						},
					},
				},
			},
			// ACTIVITY LOG PANEL
			GroupBox{
				Title:  "Activity Log",
				Layout: VBox{},
				Children: []Widget{
					TextEdit{
						AssignTo: &logText,
						ReadOnly: true,
						VScroll:  true,
					},
				},
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	go checkUpdatesWindows(app, mw)

	mw.Run()
}

func checkUpdatesWindows(coreApp *core.App, mw *walk.MainWindow) {
	info, hasUpdate, err := core.CheckUpdate("")
	if err != nil {
		log.Printf("[gui/windows] Failed to check for updates: %v", err)
		return
	}
	if hasUpdate && info != nil {
		mw.Synchronize(func() {
			msg := fmt.Sprintf("A new version (%s) is available!\r\n\r\nRelease Notes:\r\n%s\r\n\r\nWould you like to download and install it now?", info.Version, info.ReleaseNotes)
			result := walk.MsgBox(mw, "Update Available", msg, walk.MsgBoxYesNo|walk.MsgBoxIconQuestion)
			if result == walk.DlgCmdYes {
				log.Printf("[gui/windows] User accepted update to %s", info.Version)
				go performUpdateWindows(info.WindowsURL, mw)
			}
		})
	}
}

func performUpdateWindows(url string, mw *walk.MainWindow) {
	err := core.ApplyUpdate(url)
	mw.Synchronize(func() {
		if err != nil {
			walk.MsgBox(mw, "Update Failed", fmt.Sprintf("Error applying update: %v", err), walk.MsgBoxIconError)
		} else {
			walk.MsgBox(mw, "Update Successful", "Update successful! Please restart the application (it will exit now).", walk.MsgBoxIconInformation)
			os.Exit(0)
		}
	})
}
