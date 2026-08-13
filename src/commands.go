package main

import (
	"errors"
	"net"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

const offlineMessage = "No internet connection was detected."

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
	return exec.Command(command[0], command[1:]...).Start()
}

func launchTerminalCommand(title, shellCommand string, requireInternet bool) error {
	if requireInternet && !hasInternetConnection() {
		return errors.New(offlineMessage)
	}

	command := terminalCommand(title, shellCommand)
	if len(command) == 0 {
		return errors.New("no supported terminal emulator was found")
	}
	return launchCommand(command)
}

func terminalCommand(title, shellCommand string) []string {
	holdCommand := shellCommand + "; echo; read -rp 'Press Enter to close...'"
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

func hasInternetConnection() bool {
	conn, err := net.DialTimeout("tcp", "1.1.1.1:53", 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func countAvailableUpdates() (int, error) {
	if !hasInternetConnection() {
		return 0, errors.New(offlineMessage)
	}

	output, err := exec.Command("eopkg", "list-upgrades").CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = "unable to check updates"
		}
		return 0, errors.New(message)
	}

	count := 0
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "No packages to upgrade") {
			count++
		}
	}
	return count, nil
}
