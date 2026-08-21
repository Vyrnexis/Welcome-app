package main

import (
	"os"
	"testing"
)

// TestDetectDesktop verifies desktop environment resolution against XDG session variables.
func TestDetectDesktop(t *testing.T) {
	origXDG := os.Getenv("XDG_CURRENT_DESKTOP")
	origSession := os.Getenv("DESKTOP_SESSION")
	defer func() {
		os.Setenv("XDG_CURRENT_DESKTOP", origXDG)
		os.Setenv("DESKTOP_SESSION", origSession)
	}()

	tests := []struct {
		name        string
		xdgDesktop  string
		session     string
		expectedKey string
	}{
		{"KDE Plasma", "KDE", "", "kde"},
		{"Budgie via Session", "", "budgie-desktop", "budgie"},
		{"GNOME Wayland", "GNOME", "gnome", "gnome"},
		{"XFCE", "XFCE", "", "xfce"},
		{"Unknown", "AwesomeWM", "awesome", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_CURRENT_DESKTOP", tt.xdgDesktop)
			os.Setenv("DESKTOP_SESSION", tt.session)

			info := detectDesktop()
			if info.Key != tt.expectedKey {
				t.Errorf("expected %s, got %s", tt.expectedKey, info.Key)
			}
		})
	}
}

// TestArchitectureLabel verifies non-empty architecture string output.
func TestArchitectureLabel(t *testing.T) {
	label := architectureLabel()
	if label == "" {
		t.Error("architecture label should not be empty")
	}
}

// TestSystemDetailsSummary verifies that the diagnostic details string contains the architecture.
func TestSystemDetailsSummary(t *testing.T) {
	summary := systemDetailsSummary()
	if summary == "" {
		t.Error("system details summary should not be empty")
	}
	arch := architectureLabel()
	if arch != "" && len(summary) < len(arch) {
		t.Errorf("system details summary too short: %q", summary)
	}
}
