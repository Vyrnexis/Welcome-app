package main

import (
	"errors"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
)

type actionSpec struct {
	icon     func() fyne.Resource
	needsURL bool
	// internal actions are driven by the application itself and cannot be listed as configured welcome cards.
	internal bool
	// run reports whether the action succeeded, which is what counts towards getting-started progress.
	run func(ui *welcomeUI, card WelcomeCard) bool
}

// actions is the single registry of welcome-card actions; adding one here makes it valid in the configuration.
var actions = map[string]actionSpec{
	"updates": {
		icon: theme.DownloadIcon,
		run: func(ui *welcomeUI, card WelcomeCard) bool {
			title := lang.L(card.Title)
			if err := launchTerminalCommand(title, "sudo eopkg upgrade", lang.L(ui.content.UI.PressEnterToClose)); err != nil {
				dialog.ShowInformation(title, ui.errorText(err), ui.window)
				return false
			}
			return true
		},
	},
	"software": {
		icon: theme.ComputerIcon,
		run: func(ui *welcomeUI, card WelcomeCard) bool {
			return ui.openDesktopCommand(card, ui.content.SoftwareCommands)
		},
	},
	"customise": {
		icon: theme.ColorPaletteIcon,
		run: func(ui *welcomeUI, card WelcomeCard) bool {
			return ui.openDesktopCommand(card, ui.content.CustomiseCommands)
		},
	},
	"learn": {icon: theme.DocumentIcon, needsURL: true, run: (*welcomeUI).openCardURL},
	"settings": {
		icon: theme.SettingsIcon,
		run: func(ui *welcomeUI, card WelcomeCard) bool {
			return ui.openDesktopCommand(card, ui.content.SystemSettingsCommands)
		},
	},
	"whatsnew": {icon: theme.HistoryIcon, needsURL: true, run: (*welcomeUI).openCardURL},
	"nvidia":   {icon: theme.MediaVideoIcon, internal: true, run: (*welcomeUI).openDriverHelper},
	"donate":   {icon: theme.InfoIcon, needsURL: true, run: (*welcomeUI).openCardURL},
}

// openDesktopCommand launches the command configured for the current desktop, or the default entry.
func (ui *welcomeUI) openDesktopCommand(card WelcomeCard, commands map[string]CommandInfo) bool {
	command := commandFor(commands, ui.desktop.Key)
	if err := launchCommand(command.Command); err != nil {
		dialog.ShowInformation(lang.L(card.Title), commandErrorMessage(lang.L(command.Missing), err), ui.window)
		return false
	}
	return true
}

func (ui *welcomeUI) openCardURL(card WelcomeCard) bool {
	if err := openURL(ui.app, card.URL); err != nil {
		dialog.ShowError(err, ui.window)
		return false
	}
	return true
}

// openDriverHelper starts the hardware driver tool when it is installed and otherwise opens the help page,
// so the button is useful on systems without it.
func (ui *welcomeUI) openDriverHelper(card WelcomeCard) bool {
	if _, err := exec.LookPath(ui.content.Drivers.Command[0]); err != nil {
		return ui.openCardURL(card)
	}
	if err := launchCommand(ui.content.Drivers.Command); err != nil {
		dialog.ShowInformation(lang.L(card.Title), err.Error(), ui.window)
		return false
	}
	return true
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
