package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

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
	if minSize.Height != lbl1.MinSize().Height {
		t.Errorf("expected one compact row with height %f, got %f", lbl1.MinSize().Height, minSize.Height)
	}

	grid.Resize(fyne.NewSize(100, 50))
	if lbl1.Size().Height <= 0 {
		t.Errorf("expected object height > 0, got %f", lbl1.Size().Height)
	}

	grid.Resize(fyne.NewSize(640, 500))
	if lbl1.Size().Height != lbl1.MinSize().Height {
		t.Errorf("expected compact object height %f, got %f", lbl1.MinSize().Height, lbl1.Size().Height)
	}
}
