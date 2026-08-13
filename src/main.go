package main

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type welcomeUI struct {
	app          fyne.App
	window       fyne.Window
	baseDir      string
	desktop      DesktopInfo
	content      *fyne.Container
	updateButton *widget.Button
	navButtons   map[string]*widget.Button
	currentPage  string
}

func main() {
	a := app.NewWithID("us.getsol.SolusWelcome")
	a.Settings().SetTheme(newSolusTheme(false))
	w := a.NewWindow("Solus Welcome")

	baseDir := findBaseDir()

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
	ui.selectPage("welcome")
	w.ShowAndRun()
}

func (ui *welcomeUI) buildLayout() fyne.CanvasObject {
	sidebar := ui.buildSidebar()
	main := container.NewBorder(nil, ui.footer(), nil, nil, ui.content)
	return container.NewBorder(nil, nil, sidebar, nil, main)
}

func (ui *welcomeUI) buildSidebar() fyne.CanvasObject {
	logo := canvas.NewImageFromFile(filepath.Join(ui.baseDir, "assets", "logo.svg"))
	logo.SetMinSize(fyne.NewSize(88, 88))
	logo.FillMode = canvas.ImageFillContain

	title := headingText("Solus", 30)
	tagline := widget.NewLabel("The personal OS for personal computers")
	tagline.Wrapping = fyne.TextWrapWord
	tagline.Alignment = fyne.TextAlignCenter
	tagline.TextStyle = fyne.TextStyle{Italic: true}

	items := []fyne.CanvasObject{
		container.NewCenter(logo),
		container.NewCenter(title),
		tagline,
		widget.NewSeparator(),
	}

	for _, item := range navItems {
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

func (ui *welcomeUI) footer() fyne.CanvasObject {
	var startup *widget.Check
	startup = widget.NewCheck("Show this welcome screen on startup", func(enabled bool) {
		if err := setAutostartEnabled(enabled, ui.baseDir); err != nil {
			dialog.ShowError(err, ui.window)
			startup.SetChecked(isAutostartEnabled())
		}
	})
	startup.SetChecked(isAutostartEnabled())

	closeButton := widget.NewButton("Close", ui.window.Close)
	themeToggle := widget.NewCheck("Dark theme", func(enabled bool) {
		ui.app.Settings().SetTheme(newSolusTheme(enabled))
		ui.refreshPage()
	})

	footerControls := container.NewHBox(startup, themeToggle)
	return container.NewPadded(container.NewBorder(nil, nil, footerControls, closeButton))
}

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

func (ui *welcomeUI) refreshPage() {
	page := ui.currentPage
	if page == "" {
		page = "welcome"
	}
	ui.content.Objects = []fyne.CanvasObject{ui.page(page)}
	ui.content.Refresh()
}

func (ui *welcomeUI) page(page string) fyne.CanvasObject {
	switch page {
	case "customise":
		return ui.customisePage()
	case "support":
		return ui.linkPage("Support", "Find official help, community discussion, and issue tracking.", supportLinks, "Open")
	case "contribute":
		return ui.linkPage("Contribute", "Solus is built by people who test, document, package, develop, and support the project.", contributeLinks, "Learn more")
	default:
		return ui.welcomePage()
	}
}

func (ui *welcomeUI) welcomePage() fyne.CanvasObject {
	title := headingText("Welcome to Solus", 40)
	subtitle := widget.NewLabel("Thank you for choosing Solus. This welcome app will help you get set up, learn about your system, and find useful resources.")
	subtitle.Wrapping = fyne.TextWrapWord
	subtitle.TextStyle = fyne.TextStyle{Italic: true}

	heroText := container.NewVBox(title, subtitle, layout.NewSpacer())
	hero := responsiveGrid(360, 2, heroText, ui.systemCard())

	cards := make([]fyne.CanvasObject, 0, len(welcomeCards))
	for _, card := range welcomeCards {
		card := card
		cards = append(cards, ui.welcomeCard(card))
	}

	sectionTitle := headingText("Getting started", 24)
	page := container.NewVBox(
		hero,
		widget.NewSeparator(),
		sectionTitle,
		responsiveGrid(260, 3, cards...),
	)
	return container.NewPadded(container.NewVScroll(maxWidth(page, 1180)))
}

func (ui *welcomeUI) systemCard() fyne.CanvasObject {
	desktopIcon := desktopIcon(ui.baseDir, ui.desktop)

	name := headingText("Solus Linux", 22)
	edition := widget.NewLabel(ui.desktop.Name + " Edition")
	edition.TextStyle = fyne.TextStyle{Bold: true}
	arch := widget.NewLabel(architectureLabel())

	ui.updateButton = widget.NewButton("Check for updates", func() {
		ui.checkUpdates(true)
	})
	ui.updateButton.Importance = widget.HighImportance

	ui.checkUpdates(false)
	details := container.NewVBox(name, edition, arch, layout.NewSpacer(), ui.updateButton)
	return widget.NewCard("", "", container.NewHBox(desktopIcon, details))
}

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

func (ui *welcomeUI) customisePage() fyne.CanvasObject {
	actions, ok := desktopActions[ui.desktop.Key]
	if !ok {
		actions = []DesktopAction{
			{Title: "Open Settings", Body: "Desktop-specific settings were not detected."},
		}
	}

	cards := make([]fyne.CanvasObject, 0, len(actions))
	for _, action := range actions {
		action := action
		cards = append(cards, ui.standardCard(action.Title, action.Body, "Open", func() {
			if len(action.Command) == 0 {
				dialog.ShowInformation(action.Title, "Desktop-specific settings were not detected.", ui.window)
				return
			}
			if err := launchCommand(action.Command); err != nil {
				dialog.ShowInformation(action.Title, fmt.Sprintf("%s is not installed on this system.", action.Command[0]), ui.window)
			}
		}))
	}

	return ui.cardPage("Customise", "Desktop-specific settings and appearance actions based on the current session.", cards)
}

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

func (ui *welcomeUI) cardPage(title, subtitle string, cards []fyne.CanvasObject) fyne.CanvasObject {
	pageTitle := headingText(title, 34)
	pageSubtitle := widget.NewLabel(subtitle)
	pageSubtitle.Wrapping = fyne.TextWrapWord
	pageSubtitle.TextStyle = fyne.TextStyle{Italic: true}

	grid := responsiveGrid(320, 2, cards...)
	page := container.NewVBox(pageTitle, pageSubtitle, widget.NewSeparator(), grid)
	return container.NewPadded(container.NewVScroll(maxWidth(page, 1180)))
}

func (ui *welcomeUI) standardCard(title, body, buttonText string, callback func()) fyne.CanvasObject {
	titleLabel := headingText(title, 18)
	bodyLabel := widget.NewLabel(body)
	bodyLabel.Wrapping = fyne.TextWrapWord

	button := widget.NewButton(buttonText, callback)
	return widget.NewCard("", "", container.NewVBox(titleLabel, widget.NewSeparator(), bodyLabel, layout.NewSpacer(), button))
}

func (ui *welcomeUI) openWelcomeAction(action welcomeAction) {
	switch action {
	case actionUpdates:
		if err := launchTerminalCommand("Update your system", "sudo eopkg upgrade", true); err != nil {
			dialog.ShowInformation("Update your system", err.Error(), ui.window)
		}
	case actionSoftware:
		ui.openDesktopCommand("Install common apps", softwareCommands, CommandInfo{[]string{"plasma-discover"}, "No supported software centre was found."})
	case actionCustomise:
		ui.openDesktopCommand("Customise your desktop", customiseCommands, CommandInfo{[]string{"budgie-desktop-settings"}, "No supported desktop settings application was found."})
	case actionLearn:
		if err := openURL(ui.app, "https://help.getsol.us/docs/user/intro"); err != nil {
			dialog.ShowError(err, ui.window)
		}
	case actionSettings:
		ui.openDesktopCommand("System Settings", systemSettingsCommands, CommandInfo{[]string{"budgie-control-center"}, "No supported settings tool was found."})
	case actionDonate:
		if err := openURL(ui.app, "https://opencollective.com/getsolus"); err != nil {
			dialog.ShowError(err, ui.window)
		}
	}
}

func (ui *welcomeUI) openDesktopCommand(title string, commands map[string]CommandInfo, fallback CommandInfo) {
	command, ok := commands[ui.desktop.Key]
	if !ok {
		command = fallback
	}

	if err := launchCommand(command.Command); err != nil {
		dialog.ShowInformation(title, command.Missing, ui.window)
	}
}

func (ui *welcomeUI) checkUpdates(showErrors bool) {
	if ui.updateButton == nil {
		return
	}

	ui.updateButton.SetText("Checking updates...")
	ui.updateButton.Disable()

	go func() {
		count, err := countAvailableUpdates()
		fyne.Do(func() {
			ui.finishUpdateCheck(count, err, showErrors)
		})
	}()
}

func (ui *welcomeUI) finishUpdateCheck(count int, err error, showErrors bool) {
	if err != nil {
		ui.updateButton.SetText("Unable to check updates")
		if showErrors {
			dialog.ShowInformation("Check for updates", err.Error(), ui.window)
		}
		ui.updateButton.Enable()
		return
	}

	switch count {
	case 0:
		ui.updateButton.SetText("System is up to date")
	case 1:
		ui.updateButton.SetText("1 update available")
	default:
		ui.updateButton.SetText(fmt.Sprintf("%d updates available", count))
	}
	ui.updateButton.Enable()
}

func headingText(text string, size float32) *canvas.Text {
	heading := canvas.NewText(text, theme.Color(theme.ColorNameForeground))
	heading.TextStyle = fyne.TextStyle{Bold: true}
	heading.TextSize = size
	return heading
}
