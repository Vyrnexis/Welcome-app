package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// newTestUI builds a welcomeUI around the embedded configuration, optionally trimming the navigation.
func newTestUI(t *testing.T, navKeys ...string) *welcomeUI {
	t.Helper()
	content := testContent(t)
	if len(navKeys) > 0 {
		var items []NavItem
		for _, key := range navKeys {
			items = append(items, NavItem{Key: key, Label: content.NavLabel(key)})
		}
		content.NavItems = items
	}
	a := test.NewApp()
	ui := newWelcomeUI(a, a.NewWindow("Test"), content, DesktopInfo{Key: "unknown"})
	ui.updates.count = func() (int, error) { return 0, nil }
	ui.updates.do = func(f func()) { f() }
	ui.updates.spawn = func(f func()) { f() }
	ui.quit = func() { t.Error("quit called unexpectedly") }
	return ui
}

// TestUISelectPage verifies navigation changes the active page and highlights its button.
func TestUISelectPage(t *testing.T) {
	ui := newTestUI(t)
	_ = ui.buildLayout()

	for _, key := range []string{"welcome", "customise", "support", "contribute"} {
		ui.selectPage(key)
		if ui.currentPage != key {
			t.Errorf("current page = %q, want %q", ui.currentPage, key)
		}
		for other, button := range ui.navButtons {
			if (button.Importance == widget.HighImportance) != (other == key) {
				t.Errorf("page %q: button %q importance = %v", key, other, button.Importance)
			}
		}
		if len(ui.pageArea.Objects) != 1 {
			t.Errorf("page %q: page area has %d objects", key, len(ui.pageArea.Objects))
		}
	}
}

func typed(t *testing.T, ui *welcomeUI, shortcut *desktop.CustomShortcut) {
	t.Helper()
	typedCanvas, ok := ui.window.Canvas().(interface{ TypedShortcut(fyne.Shortcut) })
	if !ok {
		t.Fatal("test canvas does not support typed shortcuts")
	}
	typedCanvas.TypedShortcut(shortcut)
}

// TestSetupKeybindings verifies function keys follow the configured navigation order and Escape does nothing.
func TestSetupKeybindings(t *testing.T) {
	ui := newTestUI(t, "welcome", "support", "contribute")
	ui.window.SetContent(ui.buildLayout())
	ui.setupKeybindings()

	tests := []struct {
		key  fyne.KeyName
		page string
	}{
		{fyne.KeyF1, "welcome"},
		{fyne.KeyF2, "support"},
		{fyne.KeyF3, "contribute"},
		{fyne.KeyF4, "contribute"}, // customise was removed from the configuration, so there is no fourth shortcut
	}
	for _, tt := range tests {
		typed(t, ui, &desktop.CustomShortcut{KeyName: tt.key})
		if ui.currentPage != tt.page {
			t.Errorf("shortcut %s selected %q, want %q", tt.key, ui.currentPage, tt.page)
		}
	}

	typed(t, ui, &desktop.CustomShortcut{KeyName: fyne.KeyEscape})
}

// TestCtrlQQuits verifies Ctrl+Q requests quit.
func TestCtrlQQuits(t *testing.T) {
	ui := newTestUI(t)
	ui.window.SetContent(ui.buildLayout())
	quit := 0
	ui.quit = func() { quit++ }
	ui.setupKeybindings()

	typed(t, ui, &desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl})
	if quit != 1 {
		t.Errorf("quit called %d times, want 1", quit)
	}
}

// TestSetThemeModePersistsPreference verifies theme changes survive application restarts.
func TestSetThemeModePersistsPreference(t *testing.T) {
	ui := newTestUI(t)
	ui.setThemeMode(themeModeDark)

	if ui.themeMode != themeModeDark {
		t.Error("expected dark mode state")
	}
	if got := loadThemeMode(ui.app.Preferences()); got != themeModeDark {
		t.Errorf("persisted mode = %q, want dark", got)
	}
}

// TestFooterThemeSelect verifies the selector reflects the stored mode and changes it.
func TestFooterThemeSelect(t *testing.T) {
	ui := newTestUI(t)
	ui.themeMode = themeModeLight
	footer := ui.footer()

	var selector *widget.Select
	var find func(fyne.CanvasObject)
	find = func(o fyne.CanvasObject) {
		switch v := o.(type) {
		case *widget.Select:
			selector = v
		case *fyne.Container:
			for _, child := range v.Objects {
				find(child)
			}
		}
	}
	find(footer)
	if selector == nil {
		t.Fatal("theme selector not found in footer")
	}
	if selector.SelectedIndex() != 1 {
		t.Errorf("selected index = %d, want 1 (light)", selector.SelectedIndex())
	}
	selector.SetSelectedIndex(2)
	if ui.themeMode != themeModeDark {
		t.Errorf("themeMode = %q after selecting Dark", ui.themeMode)
	}
}

// TestAllPagesBuild verifies every registered page renders with an unknown desktop.
func TestAllPagesBuild(t *testing.T) {
	ui := newTestUI(t)
	for key, spec := range pages {
		if spec.build(ui) == nil {
			t.Errorf("page %q built nothing", key)
		}
	}
}
