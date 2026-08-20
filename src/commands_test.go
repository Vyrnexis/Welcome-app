package main

import (
	"testing"
)

// TestLaunchCommand validates executing local commands and returning errors on non-existent binaries.
func TestLaunchCommand(t *testing.T) {
	if err := launchCommand(nil); err == nil {
		t.Error("expected error for empty command")
	}

	if err := launchCommand([]string{"nonexistent-executable-for-testing-12345"}); err == nil {
		t.Error("expected error for non-existent executable")
	}

	if err := launchCommand([]string{"true"}); err != nil {
		t.Errorf("expected success launching 'true', got %v", err)
	}
}

// TestShellQuote validates proper single-quote shell escaping.
func TestShellQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "'hello'"},
		{"hello world", "'hello world'"},
		{"it's", `'it'\''s'`},
	}

	for _, tt := range tests {
		result := shellQuote(tt.input)
		if result != tt.expected {
			t.Errorf("shellQuote(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
