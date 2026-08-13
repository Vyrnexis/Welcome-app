package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const autostartFileName = "solus-welcome.desktop"

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

func isAutostartEnabled() bool {
	path := autostartPath()
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func setAutostartEnabled(enabled bool, baseDir string) error {
	path := autostartPath()
	if path == "" {
		return fmt.Errorf("could not determine the autostart directory")
	}

	if !enabled {
		if _, err := os.Stat(path); err == nil {
			return os.Remove(path)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(desktopFileContent(baseDir)), 0o644)
}

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

func autostartExec(baseDir string) string {
	installedLauncher := "/usr/bin/solus-welcome"
	if _, err := os.Stat(installedLauncher); err == nil {
		return installedLauncher
	}

	localBinary := filepath.Join(baseDir, "bin", "solus-welcome")
	if _, err := os.Stat(localBinary); err == nil {
		return quoteDesktopArg(localBinary)
	}

	return quoteDesktopArg(filepath.Join(baseDir, "solus-welcome"))
}

func autostartIcon(baseDir string) string {
	installedIcon := "/usr/share/solus-welcome/assets/logo.svg"
	if _, err := os.Stat(installedIcon); err == nil {
		return installedIcon
	}
	return filepath.Join(baseDir, "assets", "logo.svg")
}

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
