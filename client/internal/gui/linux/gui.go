//go:build linux

// Package linux provides the Linux native GUI for PoEAutoFilter using gotk4 (GTK4).
package linux

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// RunGUI launches the Linux GTK4 native GUI.
// This sets up the main application window and its tabs.
func RunGUI(app *core.App) {
	log.Println("[gui/linux] Launching GTK4 Application")

	defer LogPanic("RunGUI", nil) // Currently no owner window to pass

	a := gtk.NewApplication("poeautofilter", gio.ApplicationFlagsNone)
	a.ConnectActivate(func() { activate(a, app) })

	// GTK4 parses os.Args and expects only app-relevant arguments.
	// We pass only the executable name to prevent it from choking on cobra flags.
	if code := a.Run([]string{os.Args[0]}); code > 0 {
		log.Fatalf("GTK Application exited with code %d", code)
	}
}

func activate(gtkApp *gtk.Application, coreApp *core.App) {
	win := gtk.NewApplicationWindow(gtkApp)
	win.SetTitle("PoEAutoFilter Config")
	win.SetDefaultSize(800, 600)

	// Create a VBox for the Notebook and the Log panel
	vbox := gtk.NewBox(gtk.OrientationVertical, 0)

	// Create Notebook for tabs
	notebook := gtk.NewNotebook()
	notebook.SetVExpand(true)
	notebook.SetHExpand(true)

	// 1. General Tab
	generalBox := buildGeneralTab(coreApp, win)
	notebook.AppendPage(generalBox, gtk.NewLabel("General"))

	// 2. Style Library Tab
	styleBox := buildStyleTab(coreApp, win)
	notebook.AppendPage(styleBox, gtk.NewLabel("Style Library"))

	// 3. Value Tiers Tab
	tierBox := buildTierTab(coreApp, win)
	notebook.AppendPage(tierBox, gtk.NewLabel("Value Tiers"))

	vbox.Append(notebook)

	// Activity Log Panel
	logFrame := gtk.NewFrame("Activity Log")
	logBox := gtk.NewBox(gtk.OrientationVertical, 0)
	logFrame.SetChild(logBox)

	logScrolled := gtk.NewScrolledWindow()
	logScrolled.SetVExpand(true)
	logScrolled.SetSizeRequest(-1, 150) // Fix height for log

	logBuffer := gtk.NewTextBuffer(nil)
	logView := gtk.NewTextViewWithBuffer(logBuffer)
	logView.SetEditable(false)
	logView.SetCursorVisible(false)
	logView.SetWrapMode(gtk.WrapWordChar)
	logScrolled.SetChild(logView)
	logBox.Append(logScrolled)

	vbox.Append(logFrame)

	// Bind app.LogFunc to update the GTK TextView
	coreApp.LogFunc = func(msg string) {
		cleanMsg := strings.ReplaceAll(msg, "\r\n", "\n")
		// Use glib.IdleAdd to safely update UI from the bot loop goroutine
		glib.IdleAdd(func() {
			iter := logBuffer.EndIter()
			logBuffer.Insert(iter, cleanMsg)
			mark := logBuffer.CreateMark("end", logBuffer.EndIter(), false)
			logView.ScrollToMark(mark, 0.0, false, 0.0, 0.0)
		})
	}

	win.SetChild(vbox)
	win.Show()

	// Check for updates asynchronously
	go checkUpdates(coreApp, win)
}

func checkUpdates(coreApp *core.App, win *gtk.ApplicationWindow) {
	info, hasUpdate, err := core.CheckUpdate("")
	if err != nil {
		log.Printf("[gui/linux] Failed to check for updates: %v", err)
		return
	}
	if hasUpdate && info != nil {
		glib.IdleAdd(func() {
			promptUpdateDialog(&win.Window, info)
		})
	}
}

func promptUpdateDialog(parent *gtk.Window, info *core.VersionInfo) {
	msg := fmt.Sprintf("A new version (%s) is available!\n\nRelease Notes:\n%s\n\nWould you like to download and install it now?", info.Version, info.ReleaseNotes)
	dlg := gtk.NewMessageDialog(
		parent,
		gtk.DialogModal|gtk.DialogDestroyWithParent,
		gtk.MessageQuestion,
		gtk.ButtonsYesNo,
	)
	dlg.SetMarkup(msg)
	dlg.SetTitle("Update Available")

	dlg.ConnectResponse(func(response int) {
		if response == int(gtk.ResponseYes) {
			log.Printf("[gui/linux] User accepted update to %s", info.Version)
			go performUpdate(info.LinuxURL, parent)
		}
		dlg.Destroy()
	})
	dlg.Show()
}

func performUpdate(url string, parent *gtk.Window) {
	err := core.ApplyUpdate(url)
	glib.IdleAdd(func() {
		if err != nil {
			errDlg := gtk.NewMessageDialog(
				parent,
				gtk.DialogModal|gtk.DialogDestroyWithParent,
				gtk.MessageError,
				gtk.ButtonsClose,
			)
			errDlg.SetMarkup(fmt.Sprintf("Update failed: %v", err))
			errDlg.ConnectResponse(func(response int) { errDlg.Destroy() })
			errDlg.Show()
		} else {
			successDlg := gtk.NewMessageDialog(
				parent,
				gtk.DialogModal|gtk.DialogDestroyWithParent,
				gtk.MessageInfo,
				gtk.ButtonsClose,
			)
			successDlg.SetMarkup("Update successful! Please restart the application (it will exit now).")
			successDlg.ConnectResponse(func(response int) {
				successDlg.Destroy()
				os.Exit(0)
			})
			successDlg.Show()
		}
	})
}

func buildGeneralTab(coreApp *core.App, win *gtk.ApplicationWindow) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginTop(10)
	box.SetMarginBottom(10)
	box.SetMarginStart(10)
	box.SetMarginEnd(10)

	grid := gtk.NewGrid()
	grid.SetColumnSpacing(10)
	grid.SetRowSpacing(10)

	// League
	leagueLabel := gtk.NewLabel("League:")
	leagueLabel.SetHAlign(gtk.AlignEnd)
	leagueEntry := gtk.NewEntry()
	leagueEntry.SetText(coreApp.Config.League)
	leagueEntry.SetHExpand(true)
	grid.Attach(leagueLabel, 0, 0, 1, 1)
	grid.Attach(leagueEntry, 1, 0, 2, 1)

	// Base Filter
	baseLabel := gtk.NewLabel("Base Filter:")
	baseLabel.SetHAlign(gtk.AlignEnd)
	baseEntry := gtk.NewEntry()
	baseEntry.SetText(coreApp.Config.BaseFilePath)
	baseEntry.SetSensitive(false) // Read-only look
	baseEntry.SetHExpand(true)
	baseBtn := gtk.NewButtonWithLabel("Browse...")
	grid.Attach(baseLabel, 0, 1, 1, 1)
	grid.Attach(baseEntry, 1, 1, 1, 1)
	grid.Attach(baseBtn, 2, 1, 1, 1)

	// Output Filter
	outLabel := gtk.NewLabel("Output Filter:")
	outLabel.SetHAlign(gtk.AlignEnd)
	outEntry := gtk.NewEntry()
	outEntry.SetText(coreApp.Config.FilePath)
	outEntry.SetSensitive(false)
	outEntry.SetHExpand(true)
	outBtn := gtk.NewButtonWithLabel("Browse...")
	grid.Attach(outLabel, 0, 2, 1, 1)
	grid.Attach(outEntry, 1, 2, 1, 1)
	grid.Attach(outBtn, 2, 2, 1, 1)

	box.Append(grid)

	// File Dialogs
	baseBtn.ConnectClicked(func() {
		runFileChooserDialog(&win.Window, "Select Base Filter", gtk.FileChooserActionOpen, coreApp.Config.BaseFilePath, func(path string) {
			coreApp.Config.BaseFilePath = path
			baseEntry.SetText(path)
		})
	})

	outBtn.ConnectClicked(func() {
		runFileChooserDialog(&win.Window, "Select Output Filter", gtk.FileChooserActionSave, coreApp.Config.FilePath, func(path string) {
			coreApp.Config.FilePath = path
			outEntry.SetText(path)
		})
	})

	// Override Text
	overrideLabel := gtk.NewLabel("Custom Rules Override:")
	overrideLabel.SetHAlign(gtk.AlignStart)
	box.Append(overrideLabel)

	overrideScrolled := gtk.NewScrolledWindow()
	overrideScrolled.SetVExpand(true)
	overrideBuffer := gtk.NewTextBuffer(nil)
	overrideBuffer.SetText(coreApp.Config.Override)
	overrideView := gtk.NewTextViewWithBuffer(overrideBuffer)
	overrideView.SetWrapMode(gtk.WrapWordChar)
	overrideScrolled.SetChild(overrideView)
	box.Append(overrideScrolled)

	// Start Button
	startBtn := gtk.NewButtonWithLabel("Start AutoFilter")
	startBtn.SetMarginTop(10)
	startBtn.SetMarginBottom(10)
	startBtn.SetHAlign(gtk.AlignCenter)

	startBtn.ConnectClicked(func() {
		// Save config fields before running
		coreApp.Config.League = leagueEntry.Text()

		start, end := overrideBuffer.Bounds()
		coreApp.Config.Override = overrideBuffer.Text(start, end, false)

		coreApp.UpdateConfig(coreApp.Config)

		coreApp.State.Mu.Lock()
		isRunning := coreApp.State.IsRunning
		coreApp.State.Mu.Unlock()

		if isRunning {
			coreApp.StopBot()
			startBtn.SetLabel("Start AutoFilter")
		} else {
			coreApp.StartBot()
			startBtn.SetLabel("Stop AutoFilter")
		}
	})

	box.Append(startBtn)

	return box
}

// runFileChooserDialog builds a custom inline GTK Dialog containing a FileChooserWidget.
// This is safer than FileChooserNative because it avoids D-Bus portal calls,
// which crash or hang when the app is run with root/sudo privileges.
func runFileChooserDialog(parent *gtk.Window, title string, action gtk.FileChooserAction, initialPath string, onAccept func(string)) {
	dlg := gtk.NewDialogWithFlags(title, parent, gtk.DialogModal|gtk.DialogDestroyWithParent)
	dlg.SetDefaultSize(700, 500)

	chooser := gtk.NewFileChooserWidget(action)
	chooser.SetVExpand(true)
	chooser.SetHExpand(true)

	if initialPath != "" {
		f := gio.NewFileForPath(initialPath)
		chooser.SetFile(f)
	} else if action == gtk.FileChooserActionSave {
		chooser.SetCurrentName("filter.filter")
	}

	content := dlg.ContentArea()
	content.Append(chooser)

	dlg.AddButton("Cancel", int(gtk.ResponseCancel))
	var acceptLabel string
	if action == gtk.FileChooserActionSave {
		acceptLabel = "Save"
	} else {
		acceptLabel = "Open"
	}
	dlg.AddButton(acceptLabel, int(gtk.ResponseAccept))

	dlg.ConnectResponse(func(response int) {
		if response == int(gtk.ResponseAccept) {
			f := chooser.File()
			if f != nil {
				onAccept(f.Path())
			}
		}
		dlg.Destroy()
	})

	dlg.Show()
}

func buildStyleTab(coreApp *core.App, win *gtk.ApplicationWindow) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginTop(10)
	box.SetMarginBottom(10)
	box.SetMarginStart(10)
	box.SetMarginEnd(10)

	// Buttons HBox
	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 5)

	addBtn := gtk.NewButtonWithLabel("Add Style")
	editBtn := gtk.NewButtonWithLabel("Edit Selected")
	delBtn := gtk.NewButtonWithLabel("Delete Style")

	btnBox.Append(addBtn)
	btnBox.Append(editBtn)
	btnBox.Append(delBtn)
	box.Append(btnBox)

	// TreeView for Styles
	treeScrolled := gtk.NewScrolledWindow()
	treeScrolled.SetVExpand(true)
	treeScrolled.SetHExpand(true)

	treeView := gtk.NewTreeView()
	treeView.AppendColumn(createTextColumn("Name", int(StyleNameColumn)))
	treeView.AppendColumn(createTextColumn("Preview", int(StylePreviewColumn)))

	listStoreWrapper := NewStyleListStore()
	listStoreWrapper.Load(coreApp.Config.StyleLibrary)
	treeView.SetModel(listStoreWrapper.Store)
	treeScrolled.SetChild(treeView)

	box.Append(treeScrolled)

	// Button Actions
	addBtn.ConnectClicked(func() {
		newStyle := core.Style{Name: "New Style", Actions: []core.FilterAction{}}
		coreApp.Config.StyleLibrary = append(coreApp.Config.StyleLibrary, newStyle)
		listStoreWrapper.Load(coreApp.Config.StyleLibrary)
		coreApp.UpdateConfig(coreApp.Config)
	})

	editAction := func() {
		sel := treeView.Selection()
		if model, iter, ok := sel.Selected(); ok {
			path := model.Path(iter)
			idx := path.Indices()[0] // Get row index

			if idx >= 0 && idx < len(coreApp.Config.StyleLibrary) {
				style := &coreApp.Config.StyleLibrary[idx]
				// Open dialog
				RunStyleEditor(&win.Window, style, func() {
					listStoreWrapper.Load(coreApp.Config.StyleLibrary)
					coreApp.UpdateConfig(coreApp.Config)
				})
			}
		}
	}

	editBtn.ConnectClicked(editAction)
	treeView.ConnectRowActivated(func(path *gtk.TreePath, column *gtk.TreeViewColumn) {
		editAction()
	})

	delBtn.ConnectClicked(func() {
		sel := treeView.Selection()
		if model, iter, ok := sel.Selected(); ok {
			path := model.Path(iter)
			idx := path.Indices()[0]
			if idx >= 0 && idx < len(coreApp.Config.StyleLibrary) {
				coreApp.Config.StyleLibrary = append(coreApp.Config.StyleLibrary[:idx], coreApp.Config.StyleLibrary[idx+1:]...)
				listStoreWrapper.Load(coreApp.Config.StyleLibrary)
				coreApp.UpdateConfig(coreApp.Config)
			}
		}
	})

	return box
}

func buildTierTab(coreApp *core.App, win *gtk.ApplicationWindow) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginTop(10)
	box.SetMarginBottom(10)
	box.SetMarginStart(10)
	box.SetMarginEnd(10)

	// Buttons HBox
	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 5)

	addBtn := gtk.NewButtonWithLabel("Add Tier")
	editBtn := gtk.NewButtonWithLabel("Edit Selected")
	delBtn := gtk.NewButtonWithLabel("Delete Tier")

	btnBox.Append(addBtn)
	btnBox.Append(editBtn)
	btnBox.Append(delBtn)
	box.Append(btnBox)

	// TreeView for Tiers
	treeScrolled := gtk.NewScrolledWindow()
	treeScrolled.SetVExpand(true)
	treeScrolled.SetHExpand(true)

	treeView := gtk.NewTreeView()
	treeView.AppendColumn(createTextColumn("Name", int(TierNameColumn)))
	treeView.AppendColumn(createTextColumn("Value", int(TierValueColumn)))
	treeView.AppendColumn(createTextColumn("Currency", int(TierCurrencyColumn)))
	treeView.AppendColumn(createTextColumn("Style", int(TierStyleColumn)))

	listStoreWrapper := NewTierListStore()
	listStoreWrapper.Load(coreApp.Config.Tiers)
	treeView.SetModel(listStoreWrapper.Store)
	treeScrolled.SetChild(treeView)

	box.Append(treeScrolled)

	// Button Actions
	addBtn.ConnectClicked(func() {
		newTier := core.Tier{Name: "New Tier", Value: 1.0, Currency: "Chaos", StyleName: "Default"}
		coreApp.Config.Tiers = append(coreApp.Config.Tiers, newTier)
		listStoreWrapper.Load(coreApp.Config.Tiers)
		coreApp.UpdateConfig(coreApp.Config)
	})

	editAction := func() {
		sel := treeView.Selection()
		if model, iter, ok := sel.Selected(); ok {
			path := model.Path(iter)
			idx := path.Indices()[0]

			if idx >= 0 && idx < len(coreApp.Config.Tiers) {
				tier := &coreApp.Config.Tiers[idx]
				RunTierDialog(&win.Window, tier, coreApp.Config.StyleLibrary, func() {
					listStoreWrapper.Load(coreApp.Config.Tiers)
					coreApp.UpdateConfig(coreApp.Config)
				})
			}
		}
	}

	editBtn.ConnectClicked(editAction)
	treeView.ConnectRowActivated(func(path *gtk.TreePath, column *gtk.TreeViewColumn) {
		editAction()
	})

	delBtn.ConnectClicked(func() {
		sel := treeView.Selection()
		if model, iter, ok := sel.Selected(); ok {
			path := model.Path(iter)
			idx := path.Indices()[0]
			if idx >= 0 && idx < len(coreApp.Config.Tiers) {
				coreApp.Config.Tiers = append(coreApp.Config.Tiers[:idx], coreApp.Config.Tiers[idx+1:]...)
				listStoreWrapper.Load(coreApp.Config.Tiers)
				coreApp.UpdateConfig(coreApp.Config)
			}
		}
	})

	return box
}
