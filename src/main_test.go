package main

import (
	"testing"

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
