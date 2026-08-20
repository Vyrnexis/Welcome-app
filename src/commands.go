package main

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

const offlineMessage = "No internet connection was detected."

// openURL parses a raw URL string and opens it in the user's default web browser.
func openURL(app fyne.App, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	return app.OpenURL(parsed)
}

// launchCommand starts an external executable asynchronously using the specified arguments.
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

// launchTerminalCommand runs a shell command inside a terminal emulator, optionally checking for internet.
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

// terminalCommand wraps a shell command with the necessary arguments to execute inside a detected terminal emulator.
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

// executableExists checks if a specific command-line utility is available in the system's PATH.
func executableExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// shellQuote escapes a string so it can be safely passed as a single argument in a shell command.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// hasInternetConnection verifies internet connectivity by attempting a quick TCP connection to public DNS servers.
func hasInternetConnection() bool {
	targets := []string{"1.1.1.1:53", "8.8.8.8:53", "9.9.9.9:53"}
	for _, target := range targets {
		conn, err := net.DialTimeout("tcp", target, 1*time.Second)
		if err == nil {
			_ = conn.Close()
			return true
		}
	}
	return false
}

// countAvailableUpdates queries the eopkg package manager to determine the number of available system updates.
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

	return parseAvailableUpdates(output), nil
}

// parseAvailableUpdates counts the upgrade lines from eopkg list-upgrades output bytes.
func parseAvailableUpdates(output []byte) int {
	count := 0
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "No packages to upgrade") {
			count++
		}
	}
	return count
}
