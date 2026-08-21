package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetSystemLocales verifies locale normalization and candidate ordering.
func TestGetSystemLocales(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want []string
	}{
		{name: "regional locale", lang: "fr_CA.UTF-8", want: []string{"fr_CA", "fr"}},
		{name: "simple locale", lang: "de.UTF-8", want: []string{"de"}},
		{name: "locale with modifier", lang: "sr_RS@latin", want: []string{"sr_RS", "sr"}},
		{name: "C locale", lang: "C", want: []string{"en"}},
		{name: "POSIX locale", lang: "POSIX", want: []string{"en"}},
		{name: "empty locale", lang: "", want: []string{"en"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LANG", tt.lang)
			got := getSystemLocales()
			if len(got) != len(tt.want) {
				t.Fatalf("getSystemLocales() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getSystemLocales()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestLoadContent verifies base configuration loading.
func TestLoadContent(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	tmpDir := t.TempDir()
	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("create assets dir: %v", err)
	}

	config := []byte(`
[[NavItems]]
Key = "welcome"
Label = "Welcome"

[[WelcomeCards]]
Title = "Test Title"
Action = "updates"
`)
	if err := os.WriteFile(filepath.Join(assetsDir, "config.toml"), config, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := LoadContent(tmpDir); err != nil {
		t.Fatalf("LoadContent() error = %v", err)
	}

	if label := Content.NavLabel("welcome", "Fallback"); label != "Welcome" {
		t.Errorf("NavLabel(welcome) = %q, want %q", label, "Welcome")
	}
	if label := Content.NavLabel("unknown", "Fallback"); label != "Fallback" {
		t.Errorf("NavLabel(unknown) = %q, want %q", label, "Fallback")
	}
	if title := Content.WelcomeCardTitle("updates", "Fallback"); title != "Test Title" {
		t.Errorf("WelcomeCardTitle(updates) = %q, want %q", title, "Test Title")
	}
	if title := Content.WelcomeCardTitle("missing", "Fallback"); title != "Fallback" {
		t.Errorf("WelcomeCardTitle(missing) = %q, want %q", title, "Fallback")
	}
}

// TestLoadContentLocalizedOverride verifies that regional and base locale files override configuration.
func TestLoadContentLocalizedOverride(t *testing.T) {
	t.Setenv("LANG", "fr_CA.UTF-8")
	tmpDir := t.TempDir()
	localesDir := filepath.Join(tmpDir, "assets", "locales")
	if err := os.MkdirAll(localesDir, 0o755); err != nil {
		t.Fatalf("create locales directory: %v", err)
	}

	baseConfig := []byte(`
[UI]
Tagline = "Welcome"
Close = "Close"
`)
	localizedConfig := []byte(`
[UI]
Tagline = "Bienvenue"
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "config.toml"), baseConfig, 0o644); err != nil {
		t.Fatalf("write base config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localesDir, "fr.toml"), localizedConfig, 0o644); err != nil {
		t.Fatalf("write localized config: %v", err)
	}

	if err := LoadContent(tmpDir); err != nil {
		t.Fatalf("LoadContent() error = %v", err)
	}
	if Content.UI.Tagline != "Bienvenue" {
		t.Errorf("Content.UI.Tagline = %q, want %q", Content.UI.Tagline, "Bienvenue")
	}
	if Content.UI.Close != "Close" {
		t.Errorf("Content.UI.Close = %q, want %q", Content.UI.Close, "Close")
	}
}

// TestLoadContentMalformedLocaleRetainsBase ensures base configuration is preserved if a locale file has syntax errors.
func TestLoadContentMalformedLocaleRetainsBase(t *testing.T) {
	t.Setenv("LANG", "fr_FR.UTF-8")
	tmpDir := t.TempDir()
	localesDir := filepath.Join(tmpDir, "assets", "locales")
	if err := os.MkdirAll(localesDir, 0o755); err != nil {
		t.Fatalf("create locales directory: %v", err)
	}

	baseConfig := []byte(`
[UI]
Tagline = "Welcome"
Close = "Close"
`)
	malformedConfig := []byte(`
[UI
this is invalid toml = [unclosed
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "config.toml"), baseConfig, 0o644); err != nil {
		t.Fatalf("write base config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localesDir, "fr.toml"), malformedConfig, 0o644); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	if err := LoadContent(tmpDir); err != nil {
		t.Fatalf("LoadContent() returned error for malformed locale: %v", err)
	}
	if Content.UI.Tagline != "Welcome" {
		t.Errorf("Content.UI.Tagline = %q, want base fallback %q", Content.UI.Tagline, "Welcome")
	}
	if Content.UI.Close != "Close" {
		t.Errorf("Content.UI.Close = %q, want base fallback %q", Content.UI.Close, "Close")
	}
}
