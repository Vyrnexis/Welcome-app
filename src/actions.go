package main

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
)

type actionSpec struct {
	icon     func() fyne.Resource
	needsURL bool
	run      func(ui *welcomeUI, card WelcomeCard)
}

// actions is the single registry of welcome-card actions; adding one here makes it valid in the configuration.
var actions = map[string]actionSpec{
	"updates": {
		icon: theme.DownloadIcon,
		run: func(ui *welcomeUI, card WelcomeCard) {
			title := lang.L(card.Title)
			if err := launchTerminalCommand(title, "sudo eopkg upgrade", lang.L(ui.content.UI.PressEnterToClose)); err != nil {
				dialog.ShowInformation(title, ui.errorText(err), ui.window)
			}
		},
	},
	"software": {
		icon: theme.ComputerIcon,
		run: func(ui *welcomeUI, card WelcomeCard) {
			ui.openDesktopCommand(card, ui.content.SoftwareCommands)
		},
	},
	"customise": {
		icon: theme.ColorPaletteIcon,
		run: func(ui *welcomeUI, card WelcomeCard) {
			ui.openDesktopCommand(card, ui.content.CustomiseCommands)
		},
	},
	"learn": {icon: theme.DocumentIcon, needsURL: true, run: (*welcomeUI).openCardURL},
	"settings": {
		icon: theme.SettingsIcon,
		run: func(ui *welcomeUI, card WelcomeCard) {
			ui.openDesktopCommand(card, ui.content.SystemSettingsCommands)
		},
	},
	"donate": {icon: theme.InfoIcon, needsURL: true, run: (*welcomeUI).openCardURL},
}

// openDesktopCommand launches the command configured for the current desktop, or the default entry.
func (ui *welcomeUI) openDesktopCommand(card WelcomeCard, commands map[string]CommandInfo) {
	command := commandFor(commands, ui.desktop.Key)
	if err := launchCommand(command.Command); err != nil {
		dialog.ShowInformation(lang.L(card.Title), commandErrorMessage(lang.L(command.Missing), err), ui.window)
	}
}

func (ui *welcomeUI) openCardURL(card WelcomeCard) {
	if err := openURL(ui.app, card.URL); err != nil {
		dialog.ShowError(err, ui.window)
	}
}

// errorText localizes the application's own sentinel errors and passes other errors through unchanged.
func (ui *welcomeUI) errorText(err error) string {
	switch {
	case errors.Is(err, errNoTerminal):
		return lang.L(ui.content.UI.NoTerminal)
	case errors.Is(err, errNoAutostartDir):
		return lang.L(ui.content.UI.NoAutostartDir)
	default:
		return err.Error()
	}
}
