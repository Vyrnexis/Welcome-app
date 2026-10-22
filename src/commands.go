package main

import (
	"errors"
	"net/url"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
)

var errNoTerminal = errors.New("no supported terminal emulator was found")

func openURL(app fyne.App, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	return app.OpenURL(parsed)
}

func launchCommand(command []string) error {
	if len(command) == 0 {
		return errors.New("no command provided")
	}
	if _, err := exec.LookPath(command[0]); err != nil {
		return err
	}
	cmd := exec.Command(command[0], command[1:]...)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		_ = cmd.Wait()
	}()
	return nil
}

// commandErrorMessage uses a friendly configured message only when an executable cannot be found.
func commandErrorMessage(missing string, err error) string {
	if errors.Is(err, exec.ErrNotFound) && missing != "" {
		return missing
	}
	return err.Error()
}

func launchTerminalCommand(title, shellCommand, closePrompt string) error {
	command := terminalCommand(title, shellCommand, closePrompt)
	if len(command) == 0 {
		return errNoTerminal
	}
	return launchCommand(command)
}

func terminalCommand(title, shellCommand, closePrompt string) []string {
	holdCommand := shellCommand + "; echo; read -rp " + shellQuote(closePrompt)
	for _, launcher := range terminalLaunchers {
		if executableExists(launcher.executable) {
			return launcher.command(title, holdCommand)
		}
	}

	return nil
}

type terminalLauncher struct {
	executable string
	command    func(title, shellCommand string) []string
}

var terminalLaunchers = []terminalLauncher{
	{"xdg-terminal-exec", func(title, command string) []string {
		return []string{"xdg-terminal-exec", "--title=" + title, "--", "bash", "-lc", command}
	}},
	{"gnome-terminal", func(title, command string) []string {
		return []string{"gnome-terminal", "--title", title, "--", "bash", "-lc", command}
	}},
	{"ptyxis", func(title, command string) []string {
		return []string{"ptyxis", "--title=" + title, "--", "bash", "-lc", command}
	}},
	{"kgx", func(title, command string) []string {
		return []string{"kgx", "--title=" + title, "--", "bash", "-lc", command}
	}},
	{"konsole", func(title, command string) []string {
		return []string{"konsole", "-p", "tabtitle=" + title, "-e", "bash", "-lc", command}
	}},
	{"xfce4-terminal", func(title, command string) []string {
		return []string{"xfce4-terminal", "--title", title, "--command", "bash -lc " + shellQuote(command)}
	}},
	{"xterm", func(title, command string) []string {
		return []string{"xterm", "-T", title, "-e", "bash", "-lc", command}
	}},
}

func executableExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
