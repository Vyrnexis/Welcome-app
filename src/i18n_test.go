package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var formatVerb = regexp.MustCompile(`%[-+# 0-9.]*[a-zA-Z]`)

// loadCatalogues parses every translation file, failing the test on invalid JSON.
func loadCatalogues(t *testing.T) map[string]map[string]json.RawMessage {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("..", "assets", "translations", "*.json"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no translation files found: %v", err)
	}
	catalogues := make(map[string]map[string]json.RawMessage, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var catalogue map[string]json.RawMessage
		if err := json.Unmarshal(data, &catalogue); err != nil {
			t.Fatalf("%s: invalid JSON: %v", filepath.Base(path), err)
		}
		catalogues[filepath.Base(path)] = catalogue
	}
	return catalogues
}

// TestTranslationKeysAreCurrent verifies catalogues parse and contain no keys that no longer exist in the application.
func TestTranslationKeysAreCurrent(t *testing.T) {
	messages := translatableMessages(testContent(t))
	for file, catalogue := range loadCatalogues(t) {
		for key, raw := range catalogue {
			if !slices.Contains(pluralKeys, key) && !messages[key] {
				t.Errorf("%s: stale key %q does not match any configured string", file, key)
			}
			if slices.Contains(pluralKeys, key) {
				var forms map[string]string
				if err := json.Unmarshal(raw, &forms); err != nil || forms["other"] == "" || !strings.Contains(forms["other"], "{{.Count}}") {
					t.Errorf("%s: %q needs plural forms containing {{.Count}}: %v", file, key, err)
				}
				continue
			}
			var translated string
			if err := json.Unmarshal(raw, &translated); err != nil || strings.TrimSpace(translated) == "" {
				t.Errorf("%s: %q must be a non-empty string", file, key)
				continue
			}
			if got, want := formatVerb.FindAllString(translated, -1), formatVerb.FindAllString(key, -1); strings.Join(sortedCopy(got), ",") != strings.Join(sortedCopy(want), ",") {
				t.Errorf("%s: %q placeholders %v differ from English %v", file, key, got, want)
			}
		}
	}
}

func sortedCopy(in []string) []string {
	out := slices.Clone(in)
	slices.Sort(out)
	return out
}

// TestTranslatableMessagesSkipIdentifiers verifies commands, URLs, and keys are never offered for translation.
func TestTranslatableMessagesSkipIdentifiers(t *testing.T) {
	content := testContent(t)
	messages := translatableMessages(content)
	for _, hidden := range []string{"gnome-software", "updates", "https://opencollective.com/getsolus", "https://help.getsol.us/"} {
		if messages[hidden] {
			t.Errorf("%q must not be translatable", hidden)
		}
	}
	for _, shown := range []string{"Update now", "Solus", "GNOME Software is not installed.", content.UI.NotInstalled} {
		if !messages[shown] {
			t.Errorf("%q should be translatable", shown)
		}
	}
}

// TestUpdatesAvailablePlural verifies singular and plural forms differ and carry the count.
func TestUpdatesAvailablePlural(t *testing.T) {
	one, many := updatesAvailable(1), updatesAvailable(5)
	if one == many || !strings.Contains(one, "1") || !strings.Contains(many, "5") || strings.Contains(one, "{{") {
		t.Errorf("updatesAvailable(1) = %q, updatesAvailable(5) = %q", one, many)
	}
}
