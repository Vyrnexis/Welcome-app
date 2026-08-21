package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

const updateCheckTimeout = 30 * time.Second

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

// launchTerminalCommand runs a shell command inside a supported terminal emulator.
func launchTerminalCommand(title, shellCommand string) error {
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

// countAvailableUpdates queries the eopkg package manager to determine the number of available system updates.
func countAvailableUpdates() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
	defer cancel()

	output, err := exec.CommandContext(ctx, "eopkg", "list-upgrades").CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return 0, errors.New("update check timed out")
		}
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
