package main

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type welcomeUI struct {
	app               fyne.App
	window            fyne.Window
	baseDir           string
	desktop           DesktopInfo
	content           *fyne.Container
	updateButton      *widget.Button
	navButtons        map[string]*widget.Button
	currentPage       string
	checkingUpdates   bool
	cachedUpdateCount int
	hasCheckedUpdates bool
}

// main initializes the application, sets the theme, and launches the main window.
func main() {
	a := app.NewWithID("us.getsol.SolusWelcome")
	a.Settings().SetTheme(newSolusTheme(false))
	w := a.NewWindow("Solus Welcome")

	baseDir := findBaseDir()

	if err := LoadContent(baseDir); err != nil {
		fmt.Println("Warning: Failed to load config.toml:", err)
	}

	ui := &welcomeUI{
		app:        a,
		window:     w,
		baseDir:    baseDir,
		desktop:    detectDesktop(),
		content:    container.NewStack(),
		navButtons: make(map[string]*widget.Button),
	}

	w.SetContent(ui.buildLayout())
	w.Resize(fyne.NewSize(1080, 720))
	ui.setupKeybindings()
	ui.selectPage("welcome")
	w.ShowAndRun()
}

// buildLayout constructs the primary application layout combining the sidebar and main content area.
func (ui *welcomeUI) buildLayout() fyne.CanvasObject {
	sidebar := ui.buildSidebar()
	main := container.NewBorder(nil, ui.footer(), nil, nil, ui.content)
	return container.NewBorder(nil, nil, sidebar, nil, main)
}

// buildSidebar generates the navigation sidebar containing the logo, title, and page buttons.
func (ui *welcomeUI) buildSidebar() fyne.CanvasObject {
	logo := canvas.NewImageFromFile(filepath.Join(ui.baseDir, "assets", "logo.svg"))
	logo.SetMinSize(fyne.NewSize(88, 88))
	logo.FillMode = canvas.ImageFillContain

	title := headingText("Solus", 30)
	tagline := widget.NewLabel(Content.UI.Tagline)
	tagline.Wrapping = fyne.TextWrapWord
	tagline.Alignment = fyne.TextAlignCenter
	tagline.TextStyle = fyne.TextStyle{Italic: true}

	items := []fyne.CanvasObject{
		container.NewCenter(logo),
		container.NewCenter(title),
		tagline,
		widget.NewSeparator(),
	}

	for _, item := range Content.NavItems {
		item := item
		button := widget.NewButtonWithIcon(item.Label, navIcon(item.Key), func() {
			ui.selectPage(item.Key)
		})
		button.Alignment = widget.ButtonAlignLeading
		ui.navButtons[item.Key] = button
		items = append(items, button)
	}

	items = append(items, layout.NewSpacer())
	return container.NewPadded(minSize(container.NewVBox(items...), fyne.NewSize(270, 0)))
}

// footer creates the bottom controls section containing the autostart and theme toggles.
func (ui *welcomeUI) footer() fyne.CanvasObject {
	var startup *widget.Check
	startup = widget.NewCheck(Content.UI.ShowOnStartup, func(enabled bool) {
		if err := setAutostartEnabled(enabled, ui.baseDir); err != nil {
			dialog.ShowError(err, ui.window)
			startup.SetChecked(isAutostartEnabled())
		}
	})
	startup.SetChecked(isAutostartEnabled())

	closeButton := widget.NewButton(Content.UI.Close, ui.window.Close)
	themeToggle := widget.NewCheck(Content.UI.DarkTheme, func(enabled bool) {
		ui.app.Settings().SetTheme(newSolusTheme(enabled))
		ui.refreshPage()
	})

	footerControls := container.NewHBox(startup, themeToggle)
	return container.NewPadded(container.NewBorder(nil, nil, footerControls, closeButton))
}

// selectPage switches the main content area to the specified page and updates button states.
func (ui *welcomeUI) selectPage(page string) {
	ui.currentPage = page
	for key, button := range ui.navButtons {
		if key == page {
			button.Importance = widget.HighImportance
		} else {
			button.Importance = widget.MediumImportance
		}
		button.Refresh()
	}

	ui.content.Objects = []fyne.CanvasObject{ui.page(page)}
	ui.content.Refresh()
}

// refreshPage reloads the currently active page to reflect state or theme changes.
func (ui *welcomeUI) refreshPage() {
	page := ui.currentPage
	if page == "" {
		page = "welcome"
	}
	ui.content.Objects = []fyne.CanvasObject{ui.page(page)}
	ui.content.Refresh()
}

// page maps a string identifier to the corresponding page layout object.
func (ui *welcomeUI) page(page string) fyne.CanvasObject {
	switch page {
	case "customise":
		return ui.customisePage()
	case "support":
		return ui.linkPage(Content.NavLabel("support", "Support"), "", Content.SupportLinks, Content.UI.OpenButton)
	case "contribute":
		return ui.linkPage(Content.NavLabel("contribute", "Contribute"), "", Content.ContributeLinks, Content.UI.LearnMoreButton)
	default:
		return ui.welcomePage()
	}
}

// welcomePage builds the main hero page displaying system information and quick action cards.
func (ui *welcomeUI) welcomePage() fyne.CanvasObject {
	title := headingText(Content.UI.WelcomeTitle, 40)
	subtitle := widget.NewLabel(Content.UI.WelcomeSubtitle)
	subtitle.Wrapping = fyne.TextWrapWord
	subtitle.TextStyle = fyne.TextStyle{Italic: true}

	heroText := container.NewVBox(title, subtitle, layout.NewSpacer())
	hero := responsiveGrid(360, 2, heroText, ui.systemCard())

	cards := make([]fyne.CanvasObject, 0, len(Content.WelcomeCards))
	for _, card := range Content.WelcomeCards {
		card := card
		cards = append(cards, ui.welcomeCard(card))
	}

	sectionTitle := headingText(Content.UI.GettingStarted, 24)
	page := container.NewVBox(
		hero,
		widget.NewSeparator(),
		sectionTitle,
		responsiveGrid(260, 3, cards...),
	)
	return container.NewPadded(container.NewVScroll(maxWidth(page, 1180)))
}

// systemCard generates a detailed card displaying the current desktop edition and system diagnostics.
func (ui *welcomeUI) systemCard() fyne.CanvasObject {
	desktopIcon := desktopIcon(ui.baseDir, ui.desktop)

	name := headingText(Content.UI.SolusLinux, 22)
	edition := widget.NewLabel(ui.desktop.Name + Content.UI.EditionSuffix)
	edition.TextStyle = fyne.TextStyle{Bold: true}
	sysDetails := widget.NewLabel(systemDetailsSummary())

	ui.updateButton = widget.NewButton(Content.UI.CheckForUpdates, func() {
		ui.checkUpdates(true, true)
	})
	ui.updateButton.Importance = widget.HighImportance

	ui.checkUpdates(false, false)
	details := container.NewVBox(name, edition, sysDetails, layout.NewSpacer(), ui.updateButton)
	return widget.NewCard("", "", container.NewHBox(desktopIcon, details))
}

// setupKeybindings binds global keyboard navigation shortcuts to the application window.
func (ui *welcomeUI) setupKeybindings() {
	canvas := ui.window.Canvas()
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyEscape}, func(fyne.Shortcut) {
		ui.window.Close()
	})
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyF1}, func(fyne.Shortcut) {
		ui.selectPage("welcome")
	})
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyF2}, func(fyne.Shortcut) {
		ui.selectPage("customise")
	})
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyF3}, func(fyne.Shortcut) {
		ui.selectPage("support")
	})
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyF4}, func(fyne.Shortcut) {
		ui.selectPage("contribute")
	})
}

// welcomeCard builds a standard action card for the welcome screen grid.
func (ui *welcomeUI) welcomeCard(card WelcomeCard) fyne.CanvasObject {
	icon := widget.NewIcon(welcomeCardIcon(card.Action))
	title := headingText(card.Title, 18)
	body := widget.NewLabel(card.Body)
	body.Wrapping = fyne.TextWrapWord

	button := widget.NewButtonWithIcon(card.Button, welcomeCardIcon(card.Action), func() {
		ui.openWelcomeAction(card.Action)
	})
	button.Importance = widget.HighImportance

	return widget.NewCard("", "", container.NewVBox(
		container.NewHBox(icon, title),
		widget.NewSeparator(),
		body,
		layout.NewSpacer(),
		button,
	))
}

// customisePage constructs a grid of desktop-specific customization shortcuts.
func (ui *welcomeUI) customisePage() fyne.CanvasObject {
	actions, ok := Content.DesktopActions[ui.desktop.Key]
	if !ok {
		actions = []DesktopAction{
			{Title: Content.UI.OpenButton, Body: Content.UI.NoDesktopSettings},
		}
	}

	cards := make([]fyne.CanvasObject, 0, len(actions))
	for _, action := range actions {
		action := action
		cards = append(cards, ui.standardCard(action.Title, action.Body, Content.UI.OpenButton, func() {
			if len(action.Command) == 0 {
				dialog.ShowInformation(action.Title, Content.UI.NoDesktopSettings, ui.window)
				return
			}
			if err := launchCommand(action.Command); err != nil {
				dialog.ShowInformation(action.Title, fmt.Sprintf(Content.UI.NotInstalled, action.Command[0]), ui.window)
			}
		}))
	}

	return ui.cardPage(Content.NavLabel("customise", "Customise"), "", cards)
}

// linkPage creates a generic page layout displaying a grid of external URL link cards.
func (ui *welcomeUI) linkPage(title, subtitle string, links []LinkCard, buttonText string) fyne.CanvasObject {
	cards := make([]fyne.CanvasObject, 0, len(links))
	for _, link := range links {
		link := link
		cards = append(cards, ui.standardCard(link.Title, link.Body, buttonText, func() {
			if err := openURL(ui.app, link.URL); err != nil {
				dialog.ShowError(err, ui.window)
			}
		}))
	}
	return ui.cardPage(title, subtitle, cards)
}

// cardPage wraps a generic grid of cards with a common title and subtitle header.
func (ui *welcomeUI) cardPage(title, subtitle string, cards []fyne.CanvasObject) fyne.CanvasObject {
	pageTitle := headingText(title, 34)
	pageSubtitle := widget.NewLabel(subtitle)
	pageSubtitle.Wrapping = fyne.TextWrapWord
	pageSubtitle.TextStyle = fyne.TextStyle{Italic: true}

	grid := responsiveGrid(320, 2, cards...)
	page := container.NewVBox(pageTitle, pageSubtitle, widget.NewSeparator(), grid)
	return container.NewPadded(container.NewVScroll(maxWidth(page, 1180)))
}

// standardCard produces a uniform card layout with a title, body text, and an action button.
func (ui *welcomeUI) standardCard(title, body, buttonText string, callback func()) fyne.CanvasObject {
	titleLabel := headingText(title, 18)
	bodyLabel := widget.NewLabel(body)
	bodyLabel.Wrapping = fyne.TextWrapWord

	button := widget.NewButton(buttonText, callback)
	return widget.NewCard("", "", container.NewVBox(titleLabel, widget.NewSeparator(), bodyLabel, layout.NewSpacer(), button))
}

// openWelcomeAction executes the designated system action based on the provided action identifier string.
func (ui *welcomeUI) openWelcomeAction(action string) {
	switch action {
	case "updates":
		title := Content.WelcomeCardTitle("updates", "Update your system")
		if err := launchTerminalCommand(title, "sudo eopkg upgrade", true); err != nil {
			dialog.ShowInformation(title, err.Error(), ui.window)
		}
	case "software":
		title := Content.WelcomeCardTitle("software", "Install common apps")
		ui.openDesktopCommand(title, Content.SoftwareCommands, CommandInfo{[]string{"plasma-discover"}, "No supported software centre was found."})
	case "customise":
		title := Content.WelcomeCardTitle("customise", "Customise your desktop")
		ui.openDesktopCommand(title, Content.CustomiseCommands, CommandInfo{[]string{"budgie-desktop-settings"}, "No supported desktop settings application was found."})
	case "learn":
		if err := openURL(ui.app, "https://help.getsol.us/docs/user/intro"); err != nil {
			dialog.ShowError(err, ui.window)
		}
	case "settings":
		title := Content.WelcomeCardTitle("settings", "System Settings")
		ui.openDesktopCommand(title, Content.SystemSettingsCommands, CommandInfo{[]string{"budgie-control-center"}, "No supported settings tool was found."})
	case "donate":
		if err := openURL(ui.app, "https://opencollective.com/getsolus"); err != nil {
			dialog.ShowError(err, ui.window)
		}
	}
}

// openDesktopCommand executes a shell command from the dictionary mapped to the current desktop environment.
func (ui *welcomeUI) openDesktopCommand(title string, commands map[string]CommandInfo, fallback CommandInfo) {
	command, ok := commands[ui.desktop.Key]
	if !ok {
		command = fallback
	}

	if err := launchCommand(command.Command); err != nil {
		dialog.ShowInformation(title, command.Missing, ui.window)
	}
}

// checkUpdates triggers an asynchronous check for system package updates and updates the UI button.
func (ui *welcomeUI) checkUpdates(force, showErrors bool) {
	if ui.updateButton == nil {
		return
	}

	if !force && ui.hasCheckedUpdates {
		ui.setUpdateButtonState(ui.cachedUpdateCount)
		return
	}

	if ui.checkingUpdates {
		return
	}

	ui.checkingUpdates = true
	ui.updateButton.SetText(Content.UI.CheckingUpdates)
	ui.updateButton.Disable()

	go func() {
		count, err := countAvailableUpdates()
		fyne.Do(func() {
			ui.finishUpdateCheck(count, err, showErrors)
		})
	}()
}

// setUpdateButtonState configures the button label according to the cached update count.
func (ui *welcomeUI) setUpdateButtonState(count int) {
	if ui.updateButton == nil {
		return
	}
	switch count {
	case 0:
		ui.updateButton.SetText(Content.UI.SystemUpToDate)
	case 1:
		ui.updateButton.SetText(Content.UI.OneUpdateAvailable)
	default:
		ui.updateButton.SetText(fmt.Sprintf(Content.UI.UpdatesAvailable, count))
	}
	ui.updateButton.Enable()
}

// finishUpdateCheck updates the update button text based on the result of the package check.
func (ui *welcomeUI) finishUpdateCheck(count int, err error, showErrors bool) {
	ui.checkingUpdates = false
	if err != nil {
		ui.updateButton.SetText(Content.UI.UnableToCheckUpdates)
		if showErrors {
			dialog.ShowInformation(Content.UI.CheckForUpdates, err.Error(), ui.window)
		}
		ui.updateButton.Enable()
		return
	}

	ui.cachedUpdateCount = count
	ui.hasCheckedUpdates = true
	ui.setUpdateButtonState(count)
}

// headingText creates a styled canvas text object formatted for large header titles.
func headingText(text string, size float32) *canvas.Text {
	heading := canvas.NewText(text, theme.Color(theme.ColorNameForeground))
	heading.TextStyle = fyne.TextStyle{Bold: true}
	heading.TextSize = size
	return heading
}
