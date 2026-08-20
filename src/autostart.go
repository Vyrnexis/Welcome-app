package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const autostartFileName = "solus-welcome.desktop"

// autostartPath determines the absolute file path for the user's desktop autostart entry.
func autostartPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "autostart", autostartFileName)
}

// isAutostartEnabled checks if the welcome application is configured to launch automatically on login.
func isAutostartEnabled() bool {
	path := autostartPath()
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// setAutostartEnabled enables or disables the application autostart behavior by managing its .desktop file.
func setAutostartEnabled(enabled bool, baseDir string) error {
	path := autostartPath()
	if path == "" {
		return fmt.Errorf("could not determine the autostart directory")
	}

	if !enabled {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(desktopFileContent(baseDir)), 0o644)
}

// desktopFileContent generates the raw textual content for the autostart .desktop shortcut.
func desktopFileContent(baseDir string) string {
	return strings.Join([]string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=Solus Welcome",
		"Comment=Welcome and first-run setup for Solus",
		"Exec=" + autostartExec(baseDir),
		"Icon=" + autostartIcon(baseDir),
		"Terminal=false",
		"X-GNOME-Autostart-enabled=true",
		"",
	}, "\n")
}

// autostartExec determines the correct absolute path to the executable to use inside the autostart entry.
func autostartExec(baseDir string) string {
	installedLauncher := "/usr/bin/solus-welcome"
	if _, err := os.Stat(installedLauncher); err == nil {
		return installedLauncher
	}

	if exe, err := os.Executable(); err == nil && fileExists(exe) {
		return quoteDesktopArg(exe)
	}

	localBinary := filepath.Join(baseDir, "bin", "solus-welcome")
	if fileExists(localBinary) {
		return quoteDesktopArg(localBinary)
	}

	return quoteDesktopArg(filepath.Join(baseDir, "solus-welcome"))
}

// autostartIcon resolves the absolute path to the application logo for the desktop shortcut.
func autostartIcon(baseDir string) string {
	installedIcon := "/usr/share/solus-welcome/assets/logo.svg"
	if _, err := os.Stat(installedIcon); err == nil {
		return installedIcon
	}
	return filepath.Join(baseDir, "assets", "logo.svg")
}

// quoteDesktopArg properly escapes and quotes an executable argument for compliance with the desktop entry specification.
func quoteDesktopArg(value string) string {
	if !strings.ContainsAny(value, " \t\n\"\\$`") {
		return value
	}

	escaped := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"$", "\\$",
		"`", "\\`",
	).Replace(value)
	return `"` + escaped + `"`
}
