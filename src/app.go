package main

import (
	"errors"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type welcomeUI struct {
	app      fyne.App
	window   fyne.Window
	desktop  DesktopInfo
	host     hostPaths
	progress progress
	// rebuild redraws the current page; a field because calling render from page builders would make the pages registry initialise itself.
	rebuild     func()
	content     AppContent
	pageArea    *fyne.Container
	navButtons  map[string]*widget.Button
	currentPage string
	updates     *updateChecker
	// welcomeScroll lets rebuilds after a card action keep the user's place instead of jumping to the top.
	welcomeScroll *container.Scroll
	themeMode     themeMode
	quit          func()
	later         func(time.Duration, func()) // runs a function after a delay, off the UI goroutine
}

func main() {
	a := app.NewWithID("us.getsol.SolusWelcome")

	content, source, err := loadContent(systemConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("loaded configuration from %s", source)

	if err := migrateLegacyAutostart(); err != nil {
		log.Printf("warning: could not migrate legacy autostart entry: %v", err)
	}

	w := a.NewWindow(lang.L(content.UI.WindowTitle))
	ui := newWelcomeUI(a, w, content, detectDesktop())

	w.SetContent(ui.buildLayout())
	w.Resize(fyne.NewSize(1080, 720))
	ui.setupKeybindings()
	ui.selectPage(content.NavItems[0].Key)
	w.ShowAndRun()
}

// newWelcomeUI wires the UI state and applies the saved theme mode.
func newWelcomeUI(a fyne.App, w fyne.Window, content AppContent, desktop DesktopInfo) *welcomeUI {
	ui := &welcomeUI{
		app:        a,
		window:     w,
		desktop:    desktop,
		host:       defaultHostPaths,
		progress:   progress{a.Preferences()},
		content:    content,
		pageArea:   container.NewStack(),
		navButtons: make(map[string]*widget.Button),
		updates:    newUpdateChecker(content.UI),
		themeMode:  loadThemeMode(a.Preferences()),
		quit:       a.Quit,
		later:      func(d time.Duration, f func()) { time.AfterFunc(d, f) },
	}
	ui.updates.onError = func(err error) {
		dialog.ShowInformation(lang.L(content.UI.CheckForUpdates), ui.updates.errorText(err), w)
	}
	ui.rebuild = func() {
		var offset fyne.Position
		if ui.welcomeScroll != nil {
			offset = ui.welcomeScroll.Offset
		}
		ui.render()
		if ui.welcomeScroll != nil {
			ui.welcomeScroll.ScrollToOffset(offset)
		}
	}
	a.Settings().SetTheme(newSolusTheme(ui.themeMode))
	return ui
}

func (ui *welcomeUI) buildLayout() fyne.CanvasObject {
	main := container.NewBorder(nil, ui.footer(), nil, nil, ui.pageArea)
	return container.NewBorder(nil, nil, ui.buildSidebar(), nil, main)
}

func (ui *welcomeUI) buildSidebar() fyne.CanvasObject {
	logo := canvas.NewImageFromResource(embeddedResource("logo.svg"))
	logo.SetMinSize(fyne.NewSize(88, 88))
	logo.FillMode = canvas.ImageFillContain

	tagline := widget.NewLabel(lang.L(ui.content.UI.Tagline))
	tagline.Wrapping = fyne.TextWrapWord
	tagline.Alignment = fyne.TextAlignCenter
	tagline.TextStyle = fyne.TextStyle{Italic: true}

	items := []fyne.CanvasObject{
		container.NewCenter(logo),
		container.NewCenter(newHeading(lang.L(ui.content.UI.SidebarTitle), 30)),
		tagline,
		widget.NewSeparator(),
	}

	for _, item := range ui.content.NavItems {
		button := widget.NewButtonWithIcon(lang.L(item.Label), pages[item.Key].icon(), func() {
			ui.selectPage(item.Key)
		})
		button.Alignment = widget.ButtonAlignLeading
		ui.navButtons[item.Key] = button
		items = append(items, button)
	}

	items = append(items, layout.NewSpacer())
	return container.NewPadded(minSize(container.NewVBox(items...), fyne.NewSize(270, 0)))
}

func (ui *welcomeUI) footer() fyne.CanvasObject {
	var startup *widget.Check
	startup = widget.NewCheck(lang.L(ui.content.UI.ShowOnStartup), func(enabled bool) {
		if err := setAutostartEnabled(enabled); err != nil {
			dialog.ShowError(errors.New(ui.errorText(err)), ui.window)
			startup.SetChecked(isAutostartEnabled())
		}
	})
	startup.SetChecked(isAutostartEnabled())

	labels := map[themeMode]string{
		themeModeSystem: ui.content.UI.ThemeSystem,
		themeModeLight:  ui.content.UI.ThemeLight,
		themeModeDark:   ui.content.UI.ThemeDark,
	}
	options := make([]string, len(themeModes))
	selected := 0
	for i, mode := range themeModes {
		options[i] = lang.L(labels[mode])
		if mode == ui.themeMode {
			selected = i
		}
	}
	themeSelect := widget.NewSelect(options, nil)
	themeSelect.SetSelectedIndex(selected)
	themeSelect.OnChanged = func(string) {
		ui.setThemeMode(themes(themeSelect.SelectedIndex()))
	}

	closeButton := widget.NewButton(lang.L(ui.content.UI.Close), ui.window.Close)
	controls := container.NewHBox(startup, widget.NewLabel(lang.L(ui.content.UI.ThemeLabel)), themeSelect)
	return container.NewPadded(container.NewBorder(nil, nil, controls, closeButton))
}

// themes maps a selector index back to its mode, defaulting to System for an empty selection.
func themes(index int) themeMode {
	if index < 0 || index >= len(themeModes) {
		return themeModeSystem
	}
	return themeModes[index]
}

// setThemeMode applies and persists the theme mode; theme-aware widgets restyle themselves on the refresh.
func (ui *welcomeUI) setThemeMode(mode themeMode) {
	ui.themeMode = mode
	ui.app.Preferences().SetString(themeModePreferenceKey, string(mode))
	ui.app.Settings().SetTheme(newSolusTheme(mode))
}

func (ui *welcomeUI) selectPage(page string) {
	ui.currentPage = page
	ui.render()
}

// render shows the current page and highlights its navigation button.
func (ui *welcomeUI) render() {
	for key, button := range ui.navButtons {
		if key == ui.currentPage {
			button.Importance = widget.HighImportance
		} else {
			button.Importance = widget.MediumImportance
		}
		button.Refresh()
	}

	ui.pageArea.Objects = []fyne.CanvasObject{pages[ui.currentPage].build(ui)}
	ui.pageArea.Refresh()
}

// setupKeybindings binds Ctrl+Q to quit and F1, F2, ... to the navigation items in configuration order.
func (ui *welcomeUI) setupKeybindings() {
	c := ui.window.Canvas()
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}, func(fyne.Shortcut) {
		ui.quit()
	})

	functionKeys := []fyne.KeyName{
		fyne.KeyF1, fyne.KeyF2, fyne.KeyF3, fyne.KeyF4, fyne.KeyF5, fyne.KeyF6,
		fyne.KeyF7, fyne.KeyF8, fyne.KeyF9, fyne.KeyF10, fyne.KeyF11, fyne.KeyF12,
	}
	for i, item := range ui.content.NavItems {
		if i >= len(functionKeys) {
			break
		}
		c.AddShortcut(&desktop.CustomShortcut{KeyName: functionKeys[i]}, func(fyne.Shortcut) {
			ui.selectPage(item.Key)
		})
	}
}
