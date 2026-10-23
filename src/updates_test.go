package main

import (
	"errors"
	"testing"

	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestParseAvailableUpdates validates counting packages from eopkg output.
func TestParseAvailableUpdates(t *testing.T) {
	output := []byte(`
package1 - desc 1
package2 - desc 2
package3 - desc 3
`)
	if count := parseAvailableUpdates(output); count != 3 {
		t.Errorf("expected 3 updates, got %d", count)
	}

	noUpdates := []byte("No packages to upgrade.\n")
	if count := parseAvailableUpdates(noUpdates); count != 0 {
		t.Errorf("expected 0 updates, got %d", count)
	}
}

// newTestChecker returns a checker whose background result is released by the test and applied on the test goroutine.
func newTestChecker(t *testing.T, ui UIStrings) (checker *updateChecker, release func(int, error), apply func()) {
	t.Helper()
	test.NewApp()
	results := make(chan func() (int, error))
	queue := make(chan func(), 1)
	checker = &updateChecker{
		ui:    ui,
		count: func() (int, error) { return (<-results)() },
		do:    func(f func()) { queue <- f },
		spawn: func(f func()) { go f() },
	}
	release = func(count int, err error) {
		results <- func() (int, error) { return count, err }
	}
	apply = func() { (<-queue)() }
	return checker, release, apply
}

// TestUpdateCheckerSurvivesRebind verifies a page rebuild mid-check shows the running state, then the result, on the new button.
func TestUpdateCheckerSurvivesRebind(t *testing.T) {
	ui := testContent(t).UI
	checker, release, apply := newTestChecker(t, ui)

	first := widget.NewButton("", nil)
	checker.Bind(first)
	if first.Text != lang.L(ui.CheckForUpdates) || first.Disabled() {
		t.Fatalf("idle button = %q disabled=%v", first.Text, first.Disabled())
	}

	checker.Start(false, false)
	if first.Text != lang.L(ui.CheckingUpdates) || !first.Disabled() {
		t.Fatalf("checking button = %q disabled=%v", first.Text, first.Disabled())
	}

	second := widget.NewButton("", nil)
	checker.Bind(second)
	if second.Text != lang.L(ui.CheckingUpdates) || !second.Disabled() {
		t.Fatalf("rebound button = %q disabled=%v, want checking and disabled", second.Text, second.Disabled())
	}
	checker.Start(false, false) // must not start a second check; release below would block on a second worker

	release(3, nil)
	apply()
	if second.Text != updatesAvailable(3) || second.Disabled() {
		t.Errorf("rebound result = %q disabled=%v, want %q and enabled", second.Text, second.Disabled(), updatesAvailable(3))
	}

	third := widget.NewButton("", nil)
	checker.Bind(third)
	if third.Text != updatesAvailable(3) || third.Disabled() {
		t.Errorf("cached result = %q disabled=%v", third.Text, third.Disabled())
	}
	if third.OnTapped == nil {
		t.Error("expected Bind to wire the tap handler")
	}
}

// TestUpdateCheckerStates verifies the up-to-date, failure, interactive error, and retry paths.
func TestUpdateCheckerStates(t *testing.T) {
	ui := testContent(t).UI
	checker, release, apply := newTestChecker(t, ui)
	var reported []error
	checker.onError = func(err error) { reported = append(reported, err) }

	button := widget.NewButton("", nil)
	checker.Bind(button)

	checker.Start(false, false)
	release(0, nil)
	apply()
	if button.Text != lang.L(ui.SystemUpToDate) {
		t.Errorf("up-to-date text = %q", button.Text)
	}

	checker.Start(false, false)
	if checker.state != updateDone {
		t.Error("a non-forced start must not rerun after a completed check")
	}

	failure := errors.New("boom")
	checker.Start(true, false)
	release(0, failure)
	apply()
	if button.Text != lang.L(ui.UnableToCheckUpdates) || button.Disabled() || len(reported) != 0 {
		t.Errorf("silent failure: text=%q disabled=%v reported=%v", button.Text, button.Disabled(), reported)
	}

	checker.Start(false, false)
	if checker.state != updateFailed {
		t.Error("a non-forced start must not retry after a failure")
	}

	test.Tap(button)
	if checker.state != updateChecking {
		t.Fatal("tapping the button must start a check")
	}
	release(0, failure)
	apply()
	if len(reported) != 1 || reported[0] != failure {
		t.Errorf("interactive failure reported %v, want %v", reported, failure)
	}
}

// TestUpdateCheckerErrorText verifies sentinel errors are localized and others pass through.
func TestUpdateCheckerErrorText(t *testing.T) {
	ui := testContent(t).UI
	checker := newUpdateChecker(ui)
	if got := checker.errorText(errUpdateTimedOut); got != lang.L(ui.UpdateTimedOut) {
		t.Errorf("timeout text = %q", got)
	}
	if got := checker.errorText(errUpdateFailed); got != lang.L(ui.UpdateCheckFailed) {
		t.Errorf("failure text = %q", got)
	}
	if got := checker.errorText(errors.New("stderr text")); got != "stderr text" {
		t.Errorf("passthrough text = %q", got)
	}
}
