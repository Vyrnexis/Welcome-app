package main

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
)

// completedActionsKey stores the actions of welcome cards that were run successfully.
const completedActionsKey = "completedActions"

// progressPluralKey is the stable message ID of the "N of M done" plural.
const progressPluralKey = "getting_started_progress"

const progressPluralFallback = "{{.Count}} of {{.Total}} done"

// progress records completed getting-started actions in the app preferences.
type progress struct {
	prefs fyne.Preferences
}

func (p progress) done(action string) bool {
	return slices.Contains(p.prefs.StringList(completedActionsKey), action)
}

func (p progress) mark(action string) {
	if p.done(action) {
		return
	}
	p.prefs.SetStringList(completedActionsKey, append(p.prefs.StringList(completedActionsKey), action))
}

func (p progress) reset() {
	p.prefs.RemoveValue(completedActionsKey)
}

// count returns how many of the given cards are done.
func (p progress) count(cards []WelcomeCard) int {
	n := 0
	for _, card := range cards {
		if p.done(card.Action) {
			n++
		}
	}
	return n
}

// progressText renders the localized, pluralized "N of M done" summary.
func progressText(done, total int) string {
	return lang.XN(progressPluralKey, progressPluralFallback, done, map[string]any{"Count": done, "Total": total})
}
