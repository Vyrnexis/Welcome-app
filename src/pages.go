package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// copiedFeedback is how long the copy button shows its confirmation.
const copiedFeedback = 2 * time.Second

type pageSpec struct {
	icon  func() fyne.Resource
	build func(ui *welcomeUI) fyne.CanvasObject
}

// pages is the single registry of navigation pages; a key must be listed here to be usable in NavItems.
var pages = map[string]pageSpec{
	"welcome":   {theme.HomeIcon, (*welcomeUI).welcomePage},
	"customise": {theme.SettingsIcon, (*welcomeUI).customisePage},
	"support": {theme.HelpIcon, func(ui *welcomeUI) fyne.CanvasObject {
		return ui.linkPage("support", ui.content.SupportLinks, ui.content.UI.OpenButton)
	}},
	"contribute": {theme.ContentAddIcon, func(ui *welcomeUI) fyne.CanvasObject {
		return ui.linkPage("contribute", ui.content.ContributeLinks, ui.content.UI.LearnMoreButton)
	}},
}

// welcomePage builds the hero section and the quick action cards.
func (ui *welcomeUI) welcomePage() fyne.CanvasObject {
	title := newHeading(lang.L(ui.content.UI.WelcomeTitle), 40)
	subtitle := widget.NewLabel(lang.L(ui.content.UI.WelcomeSubtitle))
	subtitle.Wrapping = fyne.TextWrapWord
	subtitle.TextStyle = fyne.TextStyle{Italic: true}

	heroText := container.NewVBox(title, subtitle, layout.NewSpacer())
	hero := responsiveGrid(360, 2, heroText, ui.systemCard())

	welcomeCards := ui.welcomeCards()
	cards := make([]fyne.CanvasObject, 0, len(welcomeCards))
	for _, card := range welcomeCards {
		cards = append(cards, ui.welcomeCard(card))
	}

	resetButton := widget.NewButton(lang.L(ui.content.UI.ResetProgress), func() {
		ui.progress.reset()
		ui.rebuild()
	})
	resetButton.Importance = widget.LowImportance
	progressRow := container.NewHBox(widget.NewLabel(progressText(ui.progress.count(welcomeCards), len(welcomeCards))), resetButton)

	page := container.NewVBox(
		hero,
		widget.NewSeparator(),
		newHeading(lang.L(ui.content.UI.GettingStarted), 24),
		progressRow,
		responsiveGrid(260, 3, cards...),
	)
	ui.welcomeScroll = container.NewVScroll(maxWidth(page, 1180))
	return container.NewPadded(ui.welcomeScroll)
}

// welcomeCards returns the configured cards, preceded by the NVIDIA driver card when it applies.
func (ui *welcomeUI) welcomeCards() []WelcomeCard {
	if !needsNvidiaDriver(ui.host.sysfs) {
		return ui.content.WelcomeCards
	}
	d := ui.content.Drivers
	nvidia := WelcomeCard{Title: d.Title, Body: d.Body, Button: d.Button, Action: "nvidia", URL: d.URL}
	return append([]WelcomeCard{nvidia}, ui.content.WelcomeCards...)
}

// systemCard shows the desktop edition and system details with the update button.
func (ui *welcomeUI) systemCard() fyne.CanvasObject {
	name := newHeading(ui.releaseName(), 22)
	edition := widget.NewLabel(ui.desktopName() + lang.L(ui.content.UI.EditionSuffix))
	edition.TextStyle = fyne.TextStyle{Bold: true}
	sysDetails := widget.NewLabel(systemDetailsSummary(ui.content.UI))

	updateButton := widget.NewButton("", nil)
	updateButton.Importance = widget.HighImportance
	ui.updates.Bind(updateButton)
	ui.updates.Start(false, false)

	details := container.NewVBox(name, edition, sysDetails, layout.NewSpacer(), updateButton, ui.copyInfoButton())
	return widget.NewCard("", "", container.NewHBox(desktopIcon(ui.desktop), details))
}

// releaseName prefers the installed release (e.g. "Solus 4.7 Endurance") over the generic configured name.
func (ui *welcomeUI) releaseName() string {
	if name := prettyName(ui.host.osRelease); name != "" {
		return name
	}
	return lang.L(ui.content.UI.SolusLinux)
}

// copyInfoButton copies the system report and briefly confirms it. The timer fires on another goroutine and may
// outlive the page, so it only touches the button it was created with, through fyne.Do.
func (ui *welcomeUI) copyInfoButton() *widget.Button {
	label := lang.L(ui.content.UI.CopySystemInfo)
	button := widget.NewButtonWithIcon(label, theme.ContentCopyIcon(), nil)
	button.OnTapped = func() {
		ui.app.Clipboard().SetContent(ui.systemReport())
		button.SetText(lang.L(ui.content.UI.Copied))
		ui.later(copiedFeedback, func() {
			fyne.Do(func() { button.SetText(label) })
		})
	}
	return button
}

func (ui *welcomeUI) desktopName() string {
	if ui.desktop.Name == "" {
		return lang.L(ui.content.UI.UnknownDesktop)
	}
	return ui.desktop.Name
}

func (ui *welcomeUI) welcomeCard(card WelcomeCard) fyne.CanvasObject {
	icon := actions[card.Action].icon()
	headerIcon := icon
	if ui.progress.done(card.Action) {
		headerIcon = theme.ConfirmIcon()
	}
	button := widget.NewButtonWithIcon(lang.L(card.Button), icon, func() {
		if actions[card.Action].run(ui, card) {
			ui.progress.mark(card.Action)
			ui.rebuild()
		}
	})
	button.Importance = widget.HighImportance

	body := widget.NewLabel(lang.L(card.Body))
	body.Wrapping = fyne.TextWrapWord

	return widget.NewCard("", "", container.NewVBox(
		container.NewHBox(widget.NewIcon(headerIcon), newHeading(lang.L(card.Title), 18)),
		widget.NewSeparator(),
		body,
		layout.NewSpacer(),
		button,
	))
}

// customisePage shows shortcuts for the detected desktop, or a placeholder card for unsupported ones.
func (ui *welcomeUI) customisePage() fyne.CanvasObject {
	desktopActions, ok := ui.content.DesktopActions[ui.desktop.Key]
	if !ok {
		desktopActions = []DesktopAction{{Title: ui.content.UI.OpenButton, Body: ui.content.UI.NoDesktopSettings}}
	}

	cards := make([]fyne.CanvasObject, 0, len(desktopActions))
	for _, action := range desktopActions {
		title := lang.L(action.Title)
		cards = append(cards, ui.standardCard(title, lang.L(action.Body), lang.L(ui.content.UI.OpenButton), func() {
			if len(action.Command) == 0 {
				dialog.ShowInformation(title, lang.L(ui.content.UI.NoDesktopSettings), ui.window)
				return
			}
			if err := launchCommand(action.Command); err != nil {
				missing := fmt.Sprintf(lang.L(ui.content.UI.NotInstalled), action.Command[0])
				dialog.ShowInformation(title, commandErrorMessage(missing, err), ui.window)
			}
		}))
	}

	return ui.cardPage("customise", cards)
}

// linkPage shows a grid of cards that open external URLs.
func (ui *welcomeUI) linkPage(key string, links []LinkCard, buttonText string) fyne.CanvasObject {
	cards := make([]fyne.CanvasObject, 0, len(links))
	for _, link := range links {
		cards = append(cards, ui.standardCard(lang.L(link.Title), lang.L(link.Body), lang.L(buttonText), func() {
			if err := openURL(ui.app, link.URL); err != nil {
				dialog.ShowError(err, ui.window)
			}
		}))
	}
	return ui.cardPage(key, cards)
}

// cardPage wraps a grid of cards with the navigation label of the page as its title.
func (ui *welcomeUI) cardPage(key string, cards []fyne.CanvasObject) fyne.CanvasObject {
	pageTitle := newHeading(lang.L(ui.content.NavLabel(key)), 34)
	grid := responsiveGrid(320, 2, cards...)
	page := container.NewVBox(pageTitle, widget.NewSeparator(), grid)
	return container.NewPadded(container.NewVScroll(maxWidth(page, 1180)))
}

// standardCard produces a uniform card with a title, body text, and an action button; text must be localized.
func (ui *welcomeUI) standardCard(title, body, buttonText string, callback func()) fyne.CanvasObject {
	bodyLabel := widget.NewLabel(body)
	bodyLabel.Wrapping = fyne.TextWrapWord

	button := widget.NewButton(buttonText, callback)
	return widget.NewCard("", "", container.NewVBox(newHeading(title, 18), widget.NewSeparator(), bodyLabel, layout.NewSpacer(), button))
}
