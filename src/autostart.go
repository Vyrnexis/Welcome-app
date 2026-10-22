package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var errNoAutostartDir = errors.New("could not determine the autostart directory")

const (
	autostartFileName       = "us.getsol.SolusWelcome.desktop"
	legacyAutostartFileName = "solus-welcome.desktop"
	autostartIconName       = "us.getsol.SolusWelcome"
)

// userAutostartDir is where the per-user toggle lives; empty if XDG_CONFIG_HOME is unusable.
func userAutostartDir() string {
	configHome, err := os.UserConfigDir()
	if err != nil || !filepath.IsAbs(configHome) {
		return ""
	}
	return filepath.Join(configHome, "autostart")
}

func userAutostartPath() string {
	if dir := userAutostartDir(); dir != "" {
		return filepath.Join(dir, autostartFileName)
	}
	return ""
}

// systemAutostartPaths lists the installer-provided entry in every absolute XDG_CONFIG_DIRS directory.
func systemAutostartPaths() []string {
	dirs := os.Getenv("XDG_CONFIG_DIRS")
	if dirs == "" {
		dirs = "/etc/xdg"
	}
	var paths []string
	for _, dir := range filepath.SplitList(dirs) {
		if filepath.IsAbs(dir) {
			paths = append(paths, filepath.Join(dir, "autostart", autostartFileName))
		}
	}
	return paths
}

func systemAutostartExists() bool {
	for _, path := range systemAutostartPaths() {
		if fileExists(path) {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// isAutostartEnabled prefers the user's file over the system entry, as desktop environments do.
func isAutostartEnabled() bool {
	if path := userAutostartPath(); path != "" && fileExists(path) {
		return autostartEntryEnabled(path)
	}
	return systemAutostartExists()
}

// autostartEntryEnabled reports whether a desktop file is neither Hidden nor explicitly disabled for GNOME.
func autostartEntryEnabled(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	inEntry := false
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			inEntry = line == "[Desktop Entry]"
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !inEntry || !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if (key == "Hidden" && strings.EqualFold(value, "true")) ||
			(key == "X-GNOME-Autostart-enabled" && strings.EqualFold(value, "false")) {
			return false
		}
	}
	return true
}

// setAutostartEnabled writes or removes the user override so the system entry, when present, stays the source of truth.
func setAutostartEnabled(enabled bool) error {
	path := userAutostartPath()
	if path == "" {
		return errNoAutostartDir
	}

	if !enabled {
		return writeAutostartFile(path, disabledEntryContent())
	}
	if systemAutostartExists() {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	entry, err := enabledEntryContent(exe)
	if err != nil {
		return err
	}
	return writeAutostartFile(path, entry)
}

// migrateLegacyAutostart replaces the pre-system-entry user file, treating its presence as "enabled".
func migrateLegacyAutostart() error {
	dir := userAutostartDir()
	if dir == "" {
		return nil
	}
	legacy := filepath.Join(dir, legacyAutostartFileName)
	if !fileExists(legacy) {
		return nil
	}
	if !fileExists(filepath.Join(dir, autostartFileName)) {
		if err := setAutostartEnabled(true); err != nil {
			return err
		}
	}
	return os.Remove(legacy)
}

func writeAutostartFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func disabledEntryContent() string {
	return strings.Join([]string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=Solus Welcome",
		"Hidden=true",
		"",
	}, "\n")
}

// enabledEntryContent is only used when no system entry exists (development or non-package installs).
func enabledEntryContent(executable string) (string, error) {
	exec, err := quoteDesktopArg(executable)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=Solus Welcome",
		"Comment=Welcome and first-run setup for Solus",
		"Exec=" + exec,
		"Icon=" + autostartIconName,
		"Terminal=false",
		"X-GNOME-Autostart-enabled=true",
		"",
	}, "\n"), nil
}

// quoteDesktopArg quotes an Exec argument per the desktop entry specification. Backslashes are written as four
// characters and ", $ or ` as \\ plus the character because the string-value escape layer is applied before the
// quoting layer; % is doubled so field codes are never expanded.
func quoteDesktopArg(value string) (string, error) {
	if strings.ContainsAny(value, "\n\r\x00") {
		return "", errors.New("executable path contains control characters")
	}
	value = strings.ReplaceAll(value, "%", "%%")
	if !strings.ContainsAny(value, " \t\"'\\<>~|&;$*?#()`") {
		return value, nil
	}

	escaped := strings.NewReplacer(
		`\`, `\\\\`,
		`"`, `\\"`,
		"$", `\\$`,
		"`", "\\\\`",
	).Replace(value)
	return `"` + escaped + `"`, nil
}
