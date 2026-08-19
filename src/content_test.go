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
}
