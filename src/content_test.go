package main

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"testing"
)

// copyBaseConfig copies the project configuration into an isolated content test directory.
func copyBaseConfig(t *testing.T, destination string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "assets", "config.toml"))
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(destination, "config.toml"), data, 0o644); err != nil {
		t.Fatalf("write base config: %v", err)
	}
}

// TestGetSystemLocales verifies locale normalization and candidate ordering.
func TestGetSystemLocales(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	tests := []struct {
		name string
		lang string
		want []string
	}{
		{name: "regional locale", lang: "fr_CA.UTF-8", want: []string{"fr_CA", "fr"}},
		{name: "simple locale", lang: "de.UTF-8", want: []string{"de"}},
		{name: "locale with modifier", lang: "sr_RS@latin", want: []string{"sr_RS", "sr"}},
		{name: "C locale", lang: "C", want: []string{"en"}},
		{name: "C UTF-8 locale", lang: "C.UTF-8", want: []string{"en"}},
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

// TestGetSystemLocalesPrecedence verifies LC_ALL and LC_MESSAGES override LANG.
func TestGetSystemLocalesPrecedence(t *testing.T) {
	t.Setenv("LANG", "de_DE.UTF-8")
	t.Setenv("LC_MESSAGES", "fr_FR.UTF-8")
	t.Setenv("LC_ALL", "es_ES.UTF-8")
	if got := getSystemLocales()[0]; got != "es_ES" {
		t.Errorf("getSystemLocales()[0] = %q, want %q", got, "es_ES")
	}

	t.Setenv("LC_ALL", "")
	if got := getSystemLocales()[0]; got != "fr_FR" {
		t.Errorf("getSystemLocales()[0] = %q, want %q", got, "fr_FR")
	}
}

// TestLoadContent verifies base configuration loading.
func TestLoadContent(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "en_US.UTF-8")
	originalContent := Content
	t.Cleanup(func() { Content = originalContent })

	if err := LoadContent(".."); err != nil {
		t.Fatalf("LoadContent() error = %v", err)
	}
	if label := Content.NavLabel("welcome", "Fallback"); label != "Welcome" {
		t.Errorf("NavLabel(welcome) = %q, want %q", label, "Welcome")
	}
	if label := Content.NavLabel("unknown", "Fallback"); label != "Fallback" {
		t.Errorf("NavLabel(unknown) = %q, want %q", label, "Fallback")
	}
	if title := Content.WelcomeCardTitle("updates", "Fallback"); title != "Update your system" {
		t.Errorf("WelcomeCardTitle(updates) = %q, want %q", title, "Update your system")
	}
	if title := Content.WelcomeCardTitle("missing", "Fallback"); title != "Fallback" {
		t.Errorf("WelcomeCardTitle(missing) = %q, want %q", title, "Fallback")
	}
}

// TestLoadContentLocalizedOverride verifies that regional and base locale files override configuration.
func TestLoadContentLocalizedOverride(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "fr_CA.UTF-8")
	originalContent := Content
	t.Cleanup(func() { Content = originalContent })
	tmpDir := t.TempDir()
	localesDir := filepath.Join(tmpDir, "assets", "locales")
	if err := os.MkdirAll(localesDir, 0o755); err != nil {
		t.Fatalf("create locales directory: %v", err)
	}

	copyBaseConfig(t, filepath.Join(tmpDir, "assets"))
	localizedConfig := []byte("[UI]\nTagline = \"Bienvenue\"\n")
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

// TestLoadContentMalformedRegionalLocale verifies errors are reported while a base-language fallback is loaded.
func TestLoadContentMalformedRegionalLocale(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "fr_CA.UTF-8")
	originalContent := Content
	t.Cleanup(func() { Content = originalContent })
	tmpDir := t.TempDir()
	localesDir := filepath.Join(tmpDir, "assets", "locales")
	if err := os.MkdirAll(localesDir, 0o755); err != nil {
		t.Fatalf("create locales directory: %v", err)
	}

	copyBaseConfig(t, filepath.Join(tmpDir, "assets"))
	malformedConfig := []byte("[UI\nthis is invalid toml = [unclosed\n")
	localizedConfig := []byte("[UI]\nTagline = \"Bienvenue\"\n")
	if err := os.WriteFile(filepath.Join(localesDir, "fr_CA.toml"), malformedConfig, 0o644); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localesDir, "fr.toml"), localizedConfig, 0o644); err != nil {
		t.Fatalf("write fallback locale: %v", err)
	}

	if err := LoadContent(tmpDir); err == nil {
		t.Fatal("LoadContent() error = nil, want malformed regional locale error")
	}
	if Content.UI.Tagline != "Bienvenue" {
		t.Errorf("Content.UI.Tagline = %q, want language fallback %q", Content.UI.Tagline, "Bienvenue")
	}
	if Content.UI.Close != "Close" {
		t.Errorf("Content.UI.Close = %q, want base fallback %q", Content.UI.Close, "Close")
	}
}

// TestLoadContentRejectsInvalidBase verifies malformed base configuration is marked as fatal.
func TestLoadContentRejectsInvalidBase(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "en_US.UTF-8")
	tmpDir := t.TempDir()
	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("create assets directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "config.toml"), []byte("Unknown = true\n"), 0o644); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	if err := LoadContent(tmpDir); !errors.Is(err, errBaseContent) {
		t.Fatalf("LoadContent() error = %v, want fatal base content error", err)
	}
}

// TestValidateContent verifies unsafe or unusable configuration values are rejected.
func TestValidateContent(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "en_US.UTF-8")
	originalContent := Content
	t.Cleanup(func() { Content = originalContent })
	if err := LoadContent(".."); err != nil {
		t.Fatalf("LoadContent() error = %v", err)
	}
	base := Content

	tests := []struct {
		name   string
		change func(*AppContent)
	}{
		{"unknown action", func(content *AppContent) {
			content.WelcomeCards = append([]WelcomeCard(nil), content.WelcomeCards...)
			content.WelcomeCards[0].Action = "unknown"
		}},
		{"unsafe URL", func(content *AppContent) {
			content.SupportLinks = append([]LinkCard(nil), content.SupportLinks...)
			content.SupportLinks[0].URL = "file:///etc/passwd"
		}},
		{"missing format placeholder", func(content *AppContent) {
			content.UI.UpdatesAvailable = "Updates available"
		}},
		{"invalid format placeholder", func(content *AppContent) {
			content.UI.UpdatesAvailable = "%d updates for %s"
		}},
		{"empty command", func(content *AppContent) {
			content.SoftwareCommands = maps.Clone(content.SoftwareCommands)
			command := content.SoftwareCommands["gnome"]
			command.Command = nil
			content.SoftwareCommands["gnome"] = command
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := base
			tt.change(&candidate)
			if err := validateContent(candidate); err == nil {
				t.Fatal("validateContent() error = nil, want invalid configuration error")
			}
		})
	}
}
