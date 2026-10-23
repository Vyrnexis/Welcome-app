package main

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestProgressPersistence(t *testing.T) {
	a := test.NewApp()
	p := progress{a.Preferences()}
	cards := []WelcomeCard{{Action: "updates"}, {Action: "learn"}, {Action: "donate"}}

	if p.done("learn") || p.count(cards) != 0 {
		t.Fatal("fresh preferences should have no progress")
	}
	p.mark("learn")
	p.mark("learn") // marking twice must not duplicate
	p.mark("updates")
	if got := a.Preferences().StringList(completedActionsKey); len(got) != 2 {
		t.Errorf("stored %v, want two distinct actions", got)
	}
	if !(progress{a.Preferences()}).done("learn") || p.count(cards) != 2 || p.count(cards[2:]) != 0 {
		t.Errorf("count = %d", p.count(cards))
	}
	p.reset()
	if p.count(cards) != 0 {
		t.Error("reset should clear progress")
	}
}

func TestProgressText(t *testing.T) {
	if got := progressText(1, 7); got != "1 of 7 done" {
		t.Errorf("progressText(1, 7) = %q", got)
	}
	if got := progressText(0, 7); got != "0 of 7 done" {
		t.Errorf("progressText(0, 7) = %q", got)
	}
}

func findButton(obj fyne.CanvasObject, text string) *widget.Button {
	switch o := obj.(type) {
	case *widget.Button:
		if o.Text == text {
			return o
		}
	case *widget.Card:
		return findButton(o.Content, text)
	case *fyne.Container:
		for _, child := range o.Objects {
			if b := findButton(child, text); b != nil {
				return b
			}
		}
	case *container.Scroll:
		return findButton(o.Content, text)
	}
	return nil
}

// TestWelcomeCardProgress verifies a successful action is recorded and shown, a failed one is not, and Reset clears it.
func TestWelcomeCardProgress(t *testing.T) {
	ui := newTestUI(t)
	_ = ui.buildLayout()
	ui.selectPage("welcome")

	original := actions["learn"]
	t.Cleanup(func() { actions["learn"] = original })
	succeed := false
	original.run = func(*welcomeUI, WelcomeCard) bool { return succeed }
	actions["learn"] = original

	page := func() fyne.CanvasObject { return ui.pageArea.Objects[0] }
	tapLearn := func() { test.Tap(findButton(page(), "Visit Help Centre")) }

	tapLearn()
	if ui.progress.done("learn") {
		t.Fatal("a failed action must not count as done")
	}
	succeed = true
	tapLearn()
	if !ui.progress.done("learn") {
		t.Fatal("a successful action should be recorded")
	}
	if findButton(page(), "Reset") == nil {
		t.Fatal("page has no Reset button after rebuild")
	}
	test.Tap(findButton(page(), "Reset"))
	if ui.progress.done("learn") {
		t.Error("Reset should clear progress")
	}
	if !strings.Contains(progressText(ui.progress.count(ui.welcomeCards()), len(ui.welcomeCards())), "0 of") {
		t.Error("progress should read 0 after reset")
	}
}
