package main

import (
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testContent loads the embedded configuration, ignoring any system override on the machine running the tests.
func testContent(t *testing.T) AppContent {
	t.Helper()
	content, source, err := loadContent("")
	if err != nil || source != "embedded" {
		t.Fatalf("loadContent() = %q, %v", source, err)
	}
	return content
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestLoadContentEmbedded verifies the embedded configuration is valid and used without an override.
func TestLoadContentEmbedded(t *testing.T) {
	content := testContent(t)
	if label := content.NavLabel("welcome"); label != "Welcome" {
		t.Errorf("NavLabel(welcome) = %q, want %q", label, "Welcome")
	}
	if label := content.NavLabel("unknown"); label != "unknown" {
		t.Errorf("NavLabel(unknown) = %q, want the key", label)
	}

	missing := filepath.Join(t.TempDir(), "absent.toml")
	if _, source, err := loadContent(missing); err != nil || source != "embedded" {
		t.Errorf("loadContent(missing) = %q, %v; want embedded", source, err)
	}
}

// TestLoadContentOverride verifies a valid system override replaces the embedded configuration.
func TestLoadContentOverride(t *testing.T) {
	embedded := testContent(t)
	override := strings.Replace(mustEmbeddedConfig(t), `SidebarTitle = "Solus"`, `SidebarTitle = "Custom"`, 1)
	path := writeConfig(t, override)

	content, source, err := loadContent(path)
	if err != nil {
		t.Fatal(err)
	}
	if source != path || content.UI.SidebarTitle != "Custom" || embedded.UI.SidebarTitle != "Solus" {
		t.Errorf("source=%q title=%q, want override applied", source, content.UI.SidebarTitle)
	}
}

// TestLoadContentInvalidOverrideFallsBack verifies broken overrides are ignored in favour of the embedded configuration.
func TestLoadContentInvalidOverrideFallsBack(t *testing.T) {
	for name, body := range map[string]string{
		"syntax":      "[UI\nbroken = [",
		"unknown key": "Unknown = true\n",
		"incomplete":  "[[NavItems]]\nKey = \"welcome\"\nLabel = \"Welcome\"\n",
	} {
		t.Run(name, func(t *testing.T) {
			content, source, err := loadContent(writeConfig(t, body))
			if err != nil {
				t.Fatal(err)
			}
			if source != "embedded" || content.UI.SidebarTitle != "Solus" {
				t.Errorf("source = %q, want fallback to embedded", source)
			}
		})
	}
}

func mustEmbeddedConfig(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "assets", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestDecodeContentRejectsUnknownKeys verifies stray keys are fatal for a document.
func TestDecodeContentRejectsUnknownKeys(t *testing.T) {
	if _, err := decodeContent([]byte("Unknown = true\n")); err == nil {
		t.Fatal("decodeContent() error = nil, want error")
	}
}

// TestRegistriesMatchConfiguration verifies every configured page and action resolves through the registries.
func TestRegistriesMatchConfiguration(t *testing.T) {
	content := testContent(t)
	for _, item := range content.NavItems {
		spec, ok := pages[item.Key]
		if !ok || spec.icon == nil || spec.build == nil {
			t.Errorf("page %q is incomplete in the registry", item.Key)
		}
	}
	for _, card := range content.WelcomeCards {
		spec, ok := actions[card.Action]
		if !ok || spec.icon == nil || spec.run == nil {
			t.Errorf("action %q is incomplete in the registry", card.Action)
		}
	}
	for name, spec := range actions {
		if spec.icon == nil || spec.run == nil {
			t.Errorf("action %q is incomplete", name)
		}
	}
}

// TestCommandFor verifies desktop lookups use the configured default for unknown desktops.
func TestCommandFor(t *testing.T) {
	content := testContent(t)
	if got := commandFor(content.SoftwareCommands, "budgie").Command[0]; got != "gnome-software" {
		t.Errorf("budgie software command = %q, want gnome-software", got)
	}
	if got := commandFor(content.SystemSettingsCommands, "gnome").Command; len(got) != 1 || got[0] != "gnome-control-center" {
		t.Errorf("gnome settings command = %v, want plain gnome-control-center", got)
	}
	if got := commandFor(content.SoftwareCommands, "unknown"); got.Missing != content.SoftwareCommands[defaultDesktopKey].Missing {
		t.Errorf("unknown desktop did not use the default entry: %v", got)
	}
}

// TestValidateContent verifies unsafe or unusable configuration values are rejected.
func TestValidateContent(t *testing.T) {
	base := testContent(t)
	if err := validateContent(base); err != nil {
		t.Fatalf("validateContent(base) = %v", err)
	}

	tests := []struct {
		name   string
		change func(*AppContent)
	}{
		{"unknown action", func(c *AppContent) {
			c.WelcomeCards = append([]WelcomeCard(nil), c.WelcomeCards...)
			c.WelcomeCards[0].Action = "unknown"
		}},
		{"unknown page", func(c *AppContent) {
			c.NavItems = append([]NavItem(nil), c.NavItems...)
			c.NavItems[0].Key = "unknown"
		}},
		{"duplicate page", func(c *AppContent) {
			c.NavItems = append(append([]NavItem(nil), c.NavItems...), c.NavItems[0])
		}},
		{"url action without url", func(c *AppContent) {
			c.WelcomeCards = append([]WelcomeCard(nil), c.WelcomeCards...)
			for i := range c.WelcomeCards {
				if c.WelcomeCards[i].Action == "learn" {
					c.WelcomeCards[i].URL = ""
				}
			}
		}},
		{"unsafe action url", func(c *AppContent) {
			c.WelcomeCards = append([]WelcomeCard(nil), c.WelcomeCards...)
			for i := range c.WelcomeCards {
				if c.WelcomeCards[i].Action == "donate" {
					c.WelcomeCards[i].URL = "file:///etc/passwd"
				}
			}
		}},
		{"unsafe URL", func(c *AppContent) {
			c.SupportLinks = append([]LinkCard(nil), c.SupportLinks...)
			c.SupportLinks[0].URL = "file:///etc/passwd"
		}},
		{"missing UI field", func(c *AppContent) { c.UI.SidebarTitle = " " }},
		{"missing window title", func(c *AppContent) { c.UI.WindowTitle = "" }},
		{"missing format placeholder", func(c *AppContent) { c.UI.NotInstalled = "not installed" }},
		{"invalid format placeholder", func(c *AppContent) { c.UI.NotInstalled = "%d and %s" }},
		{"invalid memory format", func(c *AppContent) { c.UI.MemoryFormat = "%s %s" }},
		{"empty command", func(c *AppContent) {
			c.SoftwareCommands = maps.Clone(c.SoftwareCommands)
			command := c.SoftwareCommands["gnome"]
			command.Command = nil
			c.SoftwareCommands["gnome"] = command
		}},
		{"missing default command", func(c *AppContent) {
			c.CustomiseCommands = maps.Clone(c.CustomiseCommands)
			delete(c.CustomiseCommands, defaultDesktopKey)
		}},
		{"unknown desktop command", func(c *AppContent) {
			c.SoftwareCommands = maps.Clone(c.SoftwareCommands)
			c.SoftwareCommands["tilingwm"] = c.SoftwareCommands["gnome"]
		}},
		{"unknown desktop actions", func(c *AppContent) {
			c.DesktopActions = maps.Clone(c.DesktopActions)
			c.DesktopActions["tilingwm"] = c.DesktopActions["gnome"]
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
