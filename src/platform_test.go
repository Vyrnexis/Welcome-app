package main

import (
	"maps"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectDesktop verifies desktop environment resolution against XDG session variables.
func TestDetectDesktop(t *testing.T) {
	tests := []struct {
		name        string
		xdgDesktop  string
		session     string
		gdmSession  string
		expectedKey string
	}{
		{"KDE Plasma", "KDE", "", "", "kde"},
		{"Budgie via Session", "", "budgie-desktop", "", "budgie"},
		{"GNOME Wayland", "GNOME", "gnome", "", "gnome"},
		{"XFCE", "XFCE", "", "", "xfce"},
		{"GNOME via GDM", "", "", "gnome", "gnome"},
		{"Unknown", "AwesomeWM", "awesome", "", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CURRENT_DESKTOP", tt.xdgDesktop)
			t.Setenv("DESKTOP_SESSION", tt.session)
			t.Setenv("GDMSESSION", tt.gdmSession)

			info := detectDesktop()
			if info.Key != tt.expectedKey {
				t.Errorf("expected %s, got %s", tt.expectedKey, info.Key)
			}
		})
	}
}

// TestArchitectureLabel verifies non-empty architecture string output.
func TestArchitectureLabel(t *testing.T) {
	label := architectureLabel(testContent(t).UI)
	if label == "" {
		t.Error("architecture label should not be empty")
	}
}

// TestSystemDetailsSummary verifies that the diagnostic details string contains the architecture.
func TestSystemDetailsSummary(t *testing.T) {
	ui := testContent(t).UI
	summary := systemDetailsSummary(ui)
	if summary == "" {
		t.Error("system details summary should not be empty")
	}
	arch := architectureLabel(ui)
	if !strings.Contains(summary, arch) {
		t.Errorf("system details summary %q does not contain architecture %q", summary, arch)
	}
}

// TestParseOSRelease verifies quotes, comments, and malformed lines are handled and a missing file is empty.
func TestParseOSRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "os-release")
	writeFile(t, path, "# comment\nNAME=\"Solus\"\nPRETTY_NAME='Solus 4.7 Endurance'\nID=solus\n\nnot a pair\nVERSION= \"4.7\" \nEMPTY=\n")

	values := parseOSRelease(path)
	want := map[string]string{"NAME": "Solus", "PRETTY_NAME": "Solus 4.7 Endurance", "ID": "solus", "VERSION": "4.7", "EMPTY": ""}
	if !maps.Equal(values, want) {
		t.Errorf("parseOSRelease() = %v, want %v", values, want)
	}
	if got := prettyName(path); got != "Solus 4.7 Endurance" {
		t.Errorf("prettyName() = %q", got)
	}
	if got := prettyName(filepath.Join(t.TempDir(), "absent")); got != "" {
		t.Errorf("prettyName(missing) = %q, want empty", got)
	}
}
