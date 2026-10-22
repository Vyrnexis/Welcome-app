package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// autostartEnv isolates the user and system autostart directories and returns both.
func autostartEnv(t *testing.T) (userDir, systemDir string) {
	t.Helper()
	config := t.TempDir()
	systemRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)
	t.Setenv("XDG_CONFIG_DIRS", systemRoot)
	return filepath.Join(config, "autostart"), filepath.Join(systemRoot, "autostart")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestAutostartWithoutSystemEntry covers dev installs: off by default, a full entry when enabled, an override when disabled.
func TestAutostartWithoutSystemEntry(t *testing.T) {
	userDir, _ := autostartEnv(t)
	userFile := filepath.Join(userDir, autostartFileName)

	if isAutostartEnabled() {
		t.Error("expected autostart to be disabled without any entry")
	}

	if err := setAutostartEnabled(true); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if !isAutostartEnabled() {
		t.Error("expected autostart to be enabled")
	}
	data, err := os.ReadFile(userFile)
	if err != nil {
		t.Fatal(err)
	}
	exe, _ := os.Executable()
	want, _ := quoteDesktopArg(exe)
	for _, line := range []string{"[Desktop Entry]", "Exec=" + want, "Icon=" + autostartIconName} {
		if !strings.Contains(string(data), line) {
			t.Errorf("user entry missing %q:\n%s", line, data)
		}
	}

	if err := setAutostartEnabled(false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if isAutostartEnabled() {
		t.Error("expected autostart to be disabled")
	}
	data, _ = os.ReadFile(userFile)
	if !strings.Contains(string(data), "Hidden=true") {
		t.Errorf("disabled entry missing Hidden=true:\n%s", data)
	}
}

// TestAutostartWithSystemEntry covers installed systems: on by default, disabled by an override, re-enabled by deleting it.
func TestAutostartWithSystemEntry(t *testing.T) {
	userDir, systemDir := autostartEnv(t)
	writeFile(t, filepath.Join(systemDir, autostartFileName), "[Desktop Entry]\nType=Application\nName=Solus Welcome\nExec=solus-welcome\n")
	userFile := filepath.Join(userDir, autostartFileName)

	if !isAutostartEnabled() {
		t.Fatal("expected a new user to default to enabled")
	}

	if err := setAutostartEnabled(false); err != nil {
		t.Fatal(err)
	}
	if isAutostartEnabled() {
		t.Error("expected override to disable autostart")
	}
	data, _ := os.ReadFile(userFile)
	for _, line := range []string{"Type=Application", "Name=Solus Welcome", "Hidden=true"} {
		if !strings.Contains(string(data), line) {
			t.Errorf("override missing %q:\n%s", line, data)
		}
	}
	if strings.Contains(string(data), "Exec=") {
		t.Errorf("override must not carry an Exec line:\n%s", data)
	}

	if err := setAutostartEnabled(true); err != nil {
		t.Fatal(err)
	}
	if fileExists(userFile) {
		t.Error("enabling with a system entry must delete the user override")
	}
	if !isAutostartEnabled() {
		t.Error("expected autostart to be enabled again")
	}
	if err := setAutostartEnabled(true); err != nil {
		t.Errorf("idempotent enable: %v", err)
	}
}

// TestAutostartSystemEntryInSecondaryDir verifies every XDG_CONFIG_DIRS entry is searched.
func TestAutostartSystemEntryInSecondaryDir(t *testing.T) {
	autostartEnv(t)
	first, second := t.TempDir(), t.TempDir()
	t.Setenv("XDG_CONFIG_DIRS", first+string(os.PathListSeparator)+"relative"+string(os.PathListSeparator)+second)
	writeFile(t, filepath.Join(second, "autostart", autostartFileName), "[Desktop Entry]\n")
	if !isAutostartEnabled() {
		t.Error("expected entry in the second directory to count")
	}
}

// TestAutostartDefaultSystemDir verifies an unset XDG_CONFIG_DIRS falls back to /etc/xdg.
func TestAutostartDefaultSystemDir(t *testing.T) {
	autostartEnv(t)
	t.Setenv("XDG_CONFIG_DIRS", "")
	paths := systemAutostartPaths()
	if len(paths) != 1 || paths[0] != filepath.Join("/etc/xdg/autostart", autostartFileName) {
		t.Errorf("systemAutostartPaths() = %v", paths)
	}
}

// TestAutostartEntryDisabledForms verifies Hidden and X-GNOME-Autostart-enabled disable an entry.
func TestAutostartEntryDisabledForms(t *testing.T) {
	userDir, systemDir := autostartEnv(t)
	writeFile(t, filepath.Join(systemDir, autostartFileName), "[Desktop Entry]\n")
	userFile := filepath.Join(userDir, autostartFileName)

	tests := []struct {
		content string
		want    bool
	}{
		{"[Desktop Entry]\nType=Application\n", true},
		{"[Desktop Entry]\nHidden=true\n", false},
		{"[Desktop Entry]\nHidden=false\n", true},
		{"[Desktop Entry]\nX-GNOME-Autostart-enabled=false\n", false},
		{"[Desktop Entry]\nX-GNOME-Autostart-enabled=true\n", true},
		{"[Desktop Entry]\n[Desktop Action x]\nHidden=true\n", true},
	}
	for _, tt := range tests {
		writeFile(t, userFile, tt.content)
		if got := isAutostartEnabled(); got != tt.want {
			t.Errorf("isAutostartEnabled() = %v for %q, want %v", got, tt.content, tt.want)
		}
	}
}

// TestMigrateLegacyAutostart verifies the old user file is replaced by the new state and then deleted.
func TestMigrateLegacyAutostart(t *testing.T) {
	t.Run("with system entry", func(t *testing.T) {
		userDir, systemDir := autostartEnv(t)
		writeFile(t, filepath.Join(systemDir, autostartFileName), "[Desktop Entry]\n")
		legacy := filepath.Join(userDir, legacyAutostartFileName)
		writeFile(t, legacy, "[Desktop Entry]\nExec=/old/path\n")

		if err := migrateLegacyAutostart(); err != nil {
			t.Fatal(err)
		}
		if fileExists(legacy) || fileExists(filepath.Join(userDir, autostartFileName)) {
			t.Error("expected legacy file removed and no override written")
		}
		if !isAutostartEnabled() {
			t.Error("expected migration to leave autostart enabled")
		}
	})

	t.Run("without system entry", func(t *testing.T) {
		userDir, _ := autostartEnv(t)
		legacy := filepath.Join(userDir, legacyAutostartFileName)
		writeFile(t, legacy, "[Desktop Entry]\n")

		if err := migrateLegacyAutostart(); err != nil {
			t.Fatal(err)
		}
		if fileExists(legacy) {
			t.Error("expected legacy file removed")
		}
		if !isAutostartEnabled() {
			t.Error("expected a full user entry to keep autostart enabled")
		}
	})

	t.Run("keeps an existing choice", func(t *testing.T) {
		userDir, systemDir := autostartEnv(t)
		writeFile(t, filepath.Join(systemDir, autostartFileName), "[Desktop Entry]\n")
		writeFile(t, filepath.Join(userDir, autostartFileName), disabledEntryContent())
		writeFile(t, filepath.Join(userDir, legacyAutostartFileName), "[Desktop Entry]\n")

		if err := migrateLegacyAutostart(); err != nil {
			t.Fatal(err)
		}
		if isAutostartEnabled() {
			t.Error("migration must not override an explicit disable")
		}
	})

	t.Run("nothing to migrate", func(t *testing.T) {
		autostartEnv(t)
		if err := migrateLegacyAutostart(); err != nil {
			t.Fatal(err)
		}
	})
}

// TestAutostartRejectsRelativeConfigHome verifies autostart files cannot escape through a relative XDG path.
func TestAutostartRejectsRelativeConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "relative-config")
	if path := userAutostartPath(); path != "" {
		t.Errorf("userAutostartPath() = %q, want empty path", path)
	}
	if err := setAutostartEnabled(true); err != errNoAutostartDir {
		t.Errorf("setAutostartEnabled() error = %v, want %v", err, errNoAutostartDir)
	}
}

// TestQuoteDesktopArg validates escaping rules for desktop entry Exec command strings.
func TestQuoteDesktopArg(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/usr/bin/simple", "/usr/bin/simple"},
		{"/path with/spaces", `"/path with/spaces"`},
		{`/path"with"quotes`, `"/path\\"with\\"quotes"`},
		{"/path$with`special", `"/path\\$with\\` + "`" + `special"`},
		{`/back\slash`, `"/back\\\\slash"`},
		{"/it's", `"/it's"`},
		{"/100%/bin", "/100%%/bin"},
		{"/100% sure", `"/100%% sure"`},
	}

	for _, tt := range tests {
		result, err := quoteDesktopArg(tt.input)
		if err != nil || result != tt.expected {
			t.Errorf("quoteDesktopArg(%q) = %q, %v; expected %q", tt.input, result, err, tt.expected)
		}
	}
	if _, err := quoteDesktopArg("/bad\npath"); err == nil {
		t.Error("expected an error for a newline in the path")
	}
}
