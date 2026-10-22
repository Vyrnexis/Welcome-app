package main

import (
	"log"
	"reflect"

	"fyne.io/fyne/v2/lang"

	"github.com/Vyrnexis/welcome-app/assets"
)

// updatesPluralKey is the stable message ID of the "N updates available" plural.
const updatesPluralKey = "updates_available"

// updatesPluralFallback is the English "other" form used when no catalogue provides the message.
const updatesPluralFallback = "{{.Count}} updates available"

// Catalogues are registered at init so every test and the app see the same messages.
func init() {
	if err := lang.AddTranslationsFS(assets.FS, "translations"); err != nil {
		log.Printf("warning: translations loaded with errors: %v", err)
	}
}

// updatesAvailable renders the localized, correctly pluralized count of pending updates.
func updatesAvailable(count int) string {
	return lang.XN(updatesPluralKey, updatesPluralFallback, count, map[string]any{"Count": count})
}

// translatableMessages returns every configured string that is passed through lang.L, which is also the set
// of valid translation keys together with updatesPluralKey.
func translatableMessages(content AppContent) map[string]bool {
	messages := make(map[string]bool)
	collectMessages(reflect.ValueOf(content), messages)
	return messages
}

// collectMessages walks v and records string values, skipping fields tagged i18n:"-".
func collectMessages(v reflect.Value, out map[string]bool) {
	switch v.Kind() {
	case reflect.String:
		if v.String() != "" {
			out[v.String()] = true
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			collectMessages(v.Index(i), out)
		}
	case reflect.Map:
		for iter := v.MapRange(); iter.Next(); {
			collectMessages(iter.Value(), out)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			if v.Type().Field(i).Tag.Get("i18n") != "-" {
				collectMessages(v.Field(i), out)
			}
		}
	}
}
