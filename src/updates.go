package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

const updateCheckTimeout = 30 * time.Second

var (
	errUpdateTimedOut = errors.New("update check timed out")
	errUpdateFailed   = errors.New("unable to check updates")
)

type updateState int

const (
	updateIdle updateState = iota
	updateChecking
	updateDone
	updateFailed
)

// updateChecker owns the update-check state so it survives page rebuilds; Bind shows it on the current button.
// All methods must run on the UI goroutine.
type updateChecker struct {
	ui      UIStrings
	count   func() (int, error)
	do      func(func()) // runs a function on the UI goroutine
	spawn   func(func()) // runs the check off the UI goroutine
	onError func(error)

	state  updateState
	result int
	button *widget.Button
}

func newUpdateChecker(ui UIStrings) *updateChecker {
	return &updateChecker{ui: ui, count: countAvailableUpdates, do: fyne.Do, spawn: func(f func()) { go f() }}
}

// Bind attaches the checker to a (possibly freshly built) button and renders the current state onto it.
func (u *updateChecker) Bind(button *widget.Button) {
	u.button = button
	button.OnTapped = func() { u.Start(true, true) }
	u.render()
}

// Start runs a check in the background. Without force it only runs if no check has completed yet; interactive
// checks report failures through onError.
func (u *updateChecker) Start(force, interactive bool) {
	if u.state == updateChecking || (!force && u.state != updateIdle) {
		return
	}
	u.state = updateChecking
	u.render()

	u.spawn(func() {
		count, err := u.count()
		u.do(func() { u.finish(count, err, interactive) })
	})
}

func (u *updateChecker) finish(count int, err error, interactive bool) {
	if err != nil {
		u.state = updateFailed
		u.render()
		if interactive && u.onError != nil {
			u.onError(err)
		}
		return
	}
	u.state = updateDone
	u.result = count
	u.render()
}

func (u *updateChecker) render() {
	if u.button == nil {
		return
	}
	switch u.state {
	case updateChecking:
		u.button.SetText(lang.L(u.ui.CheckingUpdates))
		u.button.Disable()
		return
	case updateFailed:
		u.button.SetText(lang.L(u.ui.UnableToCheckUpdates))
	case updateDone:
		if u.result == 0 {
			u.button.SetText(lang.L(u.ui.SystemUpToDate))
		} else {
			u.button.SetText(updatesAvailable(u.result))
		}
	default:
		u.button.SetText(lang.L(u.ui.CheckForUpdates))
	}
	u.button.Enable()
}

// errorText turns an update-check error into a localized message.
func (u *updateChecker) errorText(err error) string {
	switch {
	case errors.Is(err, errUpdateTimedOut):
		return lang.L(u.ui.UpdateTimedOut)
	case errors.Is(err, errUpdateFailed):
		return lang.L(u.ui.UpdateCheckFailed)
	default:
		return err.Error()
	}
}

// countAvailableUpdates asks eopkg how many packages can be upgraded. Only stdout is parsed so warnings on
// stderr are not counted as packages.
func countAvailableUpdates() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
	defer cancel()

	output, err := exec.CommandContext(ctx, "eopkg", "list-upgrades", "-N").Output()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return 0, errUpdateTimedOut
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if message := strings.TrimSpace(string(exitErr.Stderr)); message != "" {
				return 0, errors.New(message)
			}
		}
		return 0, errUpdateFailed
	}

	return parseAvailableUpdates(output), nil
}

// parseAvailableUpdates counts the upgrade lines from eopkg list-upgrades output.
func parseAvailableUpdates(output []byte) int {
	count := 0
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "No packages to upgrade") {
			count++
		}
	}
	return count
}
