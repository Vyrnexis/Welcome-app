package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadContent(t *testing.T) {
	// Create a temporary directory structure matching our app
	tmpDir := t.TempDir()
	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("failed to create assets dir: %v", err)
	}

	// Write a minimal dummy config.toml
	dummyConfig := []byte(`
[[NavItems]]
Key = "test"
Label = "Test Label"

[[WelcomeCards]]
Title = "Test Card"
Action = "updates"
`)

	configPath := filepath.Join(assetsDir, "config.toml")
	if err := os.WriteFile(configPath, dummyConfig, 0644); err != nil {
		t.Fatalf("failed to write dummy config: %v", err)
	}

	// Test loading the content
	err := LoadContent(tmpDir)
	if err != nil {
		t.Fatalf("LoadContent failed: %v", err)
	}

	if len(Content.NavItems) != 1 || Content.NavItems[0].Key != "test" {
		t.Errorf("failed to parse NavItems, got: %+v", Content.NavItems)
	}

	if len(Content.WelcomeCards) != 1 || Content.WelcomeCards[0].Action != "updates" {
		t.Errorf("failed to parse WelcomeCards, got: %+v", Content.WelcomeCards)
	}

	if label := Content.NavLabel("test", "fallback"); label != "Test Label" {
		t.Errorf("expected 'Test Label', got '%s'", label)
	}
	if label := Content.NavLabel("missing", "fallback"); label != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", label)
	}

	if title := Content.WelcomeCardTitle("updates", "fallback"); title != "Test Card" {
		t.Errorf("expected 'Test Card', got '%s'", title)
	}
	if title := Content.WelcomeCardTitle("missing", "fallback"); title != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", title)
	}
}

// TestGetSystemLocales validates extraction of locale candidate tags from environment strings.
func TestGetSystemLocales(t *testing.T) {
	tests := []struct {
		lang     string
		expected []string
	}{
		{"", []string{"en"}},
		{"C", []string{"en"}},
		{"POSIX", []string{"en"}},
		{"fr_FR.UTF-8", []string{"fr_FR", "fr"}},
		{"pt_BR.UTF-8@euro", []string{"pt_BR", "pt"}},
		{"de.UTF-8", []string{"de"}},
	}

	for _, tt := range tests {
		t.Setenv("LANG", tt.lang)
		result := getSystemLocales()
		if len(result) != len(tt.expected) {
			t.Errorf("getSystemLocales(%q) length = %d, expected %d", tt.lang, len(result), len(tt.expected))
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("getSystemLocales(%q)[%d] = %q, expected %q", tt.lang, i, result[i], tt.expected[i])
			}
		}
	}
}
