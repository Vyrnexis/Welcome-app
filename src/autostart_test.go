package main

import (
	"os"
	"strings"
	"testing"
)

// TestAutostartManagement validates creating, reading, and removing autostart desktop entries.
func TestAutostartManagement(t *testing.T) {
	tmpConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpConfig)

	baseDir := t.TempDir()

	if isAutostartEnabled() {
		t.Error("expected autostart to be disabled initially")
	}

	if err := setAutostartEnabled(true, baseDir); err != nil {
		t.Fatalf("failed to enable autostart: %v", err)
	}

	if !isAutostartEnabled() {
		t.Error("expected autostart to be enabled after setting")
	}

	content, err := os.ReadFile(autostartPath())
	if err != nil {
		t.Fatalf("failed to read autostart file: %v", err)
	}
	if !strings.Contains(string(content), "[Desktop Entry]") {
		t.Errorf("autostart file missing desktop entry header: %s", string(content))
	}

	if err := setAutostartEnabled(false, baseDir); err != nil {
		t.Fatalf("failed to disable autostart: %v", err)
	}

	if isAutostartEnabled() {
		t.Error("expected autostart to be disabled after removal")
	}

	if err := setAutostartEnabled(false, baseDir); err != nil {
		t.Errorf("idempotent disable should not return error: %v", err)
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
		{`/path"with"quotes`, `"/path\"with\"quotes"`},
		{"/path$with`special", `"/path\$with\` + "`" + `special"`},
	}

	for _, tt := range tests {
		result := quoteDesktopArg(tt.input)
		if result != tt.expected {
			t.Errorf("quoteDesktopArg(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
