package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestUISelectPage verifies navigation changes the active page.
func TestUISelectPage(t *testing.T) {
	testApp := test.NewApp()
	testWindow := testApp.NewWindow("Test")

	Content.NavItems = []NavItem{
		{"welcome", "Welcome"},
		{"customise", "Customise"},
	}

	ui := &welcomeUI{
		app:        testApp,
		window:     testWindow,
		desktop:    DesktopInfo{Key: "unknown", Name: "Test Desktop"},
		content:    container.NewStack(),
		navButtons: make(map[string]*widget.Button),
	}

	_ = ui.buildLayout()

	ui.selectPage("welcome")
	if ui.currentPage != "welcome" {
		t.Errorf("Expected current page to be 'welcome', got %s", ui.currentPage)
	}

	ui.selectPage("customise")
	if ui.currentPage != "customise" {
		t.Errorf("Expected current page to be 'customise', got %s", ui.currentPage)
	}
}

// TestSetupKeybindings verifies function-key shortcuts select their corresponding pages.
func TestSetupKeybindings(t *testing.T) {
	testApp := test.NewApp()
	testWindow := testApp.NewWindow("Test")
	ui := &welcomeUI{
		app:        testApp,
		window:     testWindow,
		desktop:    DesktopInfo{Key: "unknown", Name: "Test Desktop"},
		content:    container.NewStack(),
		navButtons: make(map[string]*widget.Button),
	}
	testWindow.SetContent(ui.content)
	ui.setupKeybindings()

	typedCanvas, ok := testWindow.Canvas().(interface{ TypedShortcut(fyne.Shortcut) })
	if !ok {
		t.Fatal("test canvas does not support typed shortcuts")
	}

	tests := []struct {
		key  fyne.KeyName
		page string
	}{
		{key: fyne.KeyF1, page: "welcome"},
		{key: fyne.KeyF2, page: "customise"},
		{key: fyne.KeyF3, page: "support"},
		{key: fyne.KeyF4, page: "contribute"},
	}
	for _, tt := range tests {
		typedCanvas.TypedShortcut(&desktop.CustomShortcut{KeyName: tt.key})
		if ui.currentPage != tt.page {
			t.Errorf("shortcut %s selected %q, want %q", tt.key, ui.currentPage, tt.page)
		}
	}
}

// TestResponsiveGrid verifies responsive layouts preserve usable child dimensions.
func TestResponsiveGrid(t *testing.T) {
	lbl1 := widget.NewLabel("Item 1")
	lbl2 := widget.NewLabel("Item 2")
	lbl3 := widget.NewLabel("Item 3")

	grid := responsiveGrid(200, 3, lbl1, lbl2, lbl3)
	minSize := grid.MinSize()
	if minSize.Width < 200 {
		t.Errorf("expected min width >= 200, got %f", minSize.Width)
	}
	if minSize.Height <= 0 {
		t.Errorf("expected min height > 0, got %f", minSize.Height)
	}

	grid.Resize(fyne.NewSize(100, 50))
	if lbl1.Size().Height <= 0 {
		t.Errorf("expected object height > 0, got %f", lbl1.Size().Height)
	}
}
