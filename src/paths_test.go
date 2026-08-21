package main

import "testing"

// TestBaseDirCandidates verifies executables installed below a custom prefix can locate shared assets.
func TestBaseDirCandidates(t *testing.T) {
	candidates := baseDirCandidates("/workspace/welcome-app", "/opt/solus/bin/solus-welcome")
	want := "/opt/solus/share/solus-welcome"
	for _, candidate := range candidates {
		if candidate == want {
			return
		}
	}
	t.Errorf("baseDirCandidates() = %v, missing %q", candidates, want)
}

// TestInstallationPrefix verifies installed asset directories resolve to their filesystem prefix.
func TestInstallationPrefix(t *testing.T) {
	tests := []struct {
		baseDir string
		want    string
	}{
		{baseDir: "/usr/share/solus-welcome", want: "/usr"},
		{baseDir: "/opt/solus/share/solus-welcome", want: "/opt/solus"},
		{baseDir: "/workspace/welcome-app", want: ""},
	}

	for _, tt := range tests {
		if got := installationPrefix(tt.baseDir); got != tt.want {
			t.Errorf("installationPrefix(%q) = %q, want %q", tt.baseDir, got, tt.want)
		}
	}
}
