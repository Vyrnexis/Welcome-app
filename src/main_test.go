package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestUISelectPage(t *testing.T) {
	// Initialize a headless Fyne test app
	testApp := test.NewApp()
	testWindow := testApp.NewWindow("Test")

	// Create dummy nav items for the test since config might not be fully loaded
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

	// Better to just build the layout
	_ = ui.buildLayout()

	// Test default page selection
	ui.selectPage("welcome")
	if ui.currentPage != "welcome" {
		t.Errorf("Expected current page to be 'welcome', got %s", ui.currentPage)
	}

	// Test switching pages
	ui.selectPage("customise")
	if ui.currentPage != "customise" {
		t.Errorf("Expected current page to be 'customise', got %s", ui.currentPage)
	}
}

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
