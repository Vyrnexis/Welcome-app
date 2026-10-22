package main

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

// TestLoadThemeModeDefaultsToSystem verifies a fresh profile follows the desktop.
func TestLoadThemeModeDefaultsToSystem(t *testing.T) {
	prefs := test.NewApp().Preferences()
	if got := loadThemeMode(prefs); got != themeModeSystem {
		t.Errorf("loadThemeMode() = %q, want system", got)
	}
}

// TestLoadThemeModeMigratesLegacyBool verifies the old darkTheme bool maps to a mode and is then removed.
func TestLoadThemeModeMigratesLegacyBool(t *testing.T) {
	tests := []struct {
		legacy bool
		want   themeMode
	}{
		{true, themeModeDark},
		{false, themeModeSystem},
	}
	for _, tt := range tests {
		prefs := test.NewApp().Preferences()
		prefs.SetBool(legacyDarkThemeKey, tt.legacy)

		if got := loadThemeMode(prefs); got != tt.want {
			t.Errorf("legacy %v: loadThemeMode() = %q, want %q", tt.legacy, got, tt.want)
		}
		if hasLegacyDarkTheme(prefs) {
			t.Errorf("legacy %v: old key still present", tt.legacy)
		}
		if got := prefs.String(themeModePreferenceKey); got != string(tt.want) {
			t.Errorf("legacy %v: stored mode = %q, want %q", tt.legacy, got, tt.want)
		}
		if got := loadThemeMode(prefs); got != tt.want {
			t.Errorf("legacy %v: second load = %q, want %q", tt.legacy, got, tt.want)
		}
	}
}

// TestLoadThemeModePrefersStoredMode verifies an existing mode wins over a leftover legacy key, which is still cleaned up.
func TestLoadThemeModePrefersStoredMode(t *testing.T) {
	prefs := test.NewApp().Preferences()
	prefs.SetString(themeModePreferenceKey, string(themeModeLight))
	prefs.SetBool(legacyDarkThemeKey, true)

	if got := loadThemeMode(prefs); got != themeModeLight {
		t.Errorf("loadThemeMode() = %q, want light", got)
	}
	if hasLegacyDarkTheme(prefs) {
		t.Error("expected the legacy key to be removed")
	}

	prefs.SetString(themeModePreferenceKey, "bogus")
	if got := loadThemeMode(prefs); got != themeModeSystem {
		t.Errorf("invalid stored mode = %q, want system", got)
	}
}

// TestSolusThemeModes verifies System honours the passed variant while Light and Dark force theirs.
func TestSolusThemeModes(t *testing.T) {
	background := func(mode themeMode, variant fyne.ThemeVariant) color.Color {
		return newSolusTheme(mode).Color(theme.ColorNameBackground, variant)
	}
	light := theme.DefaultTheme().Color(theme.ColorNameBackground, theme.VariantLight)
	dark := theme.DefaultTheme().Color(theme.ColorNameBackground, theme.VariantDark)

	tests := []struct {
		mode    themeMode
		variant fyne.ThemeVariant
		want    color.Color
	}{
		{themeModeSystem, theme.VariantLight, light},
		{themeModeSystem, theme.VariantDark, dark},
		{themeModeLight, theme.VariantDark, light},
		{themeModeDark, theme.VariantLight, dark},
	}
	for _, tt := range tests {
		if got := background(tt.mode, tt.variant); got != tt.want {
			t.Errorf("mode %s variant %v: background = %v, want %v", tt.mode, tt.variant, got, tt.want)
		}
	}

	primary := func(mode themeMode, variant fyne.ThemeVariant) color.Color {
		return newSolusTheme(mode).Color(theme.ColorNamePrimary, variant)
	}
	if primary(themeModeSystem, theme.VariantDark) != primary(themeModeDark, theme.VariantLight) {
		t.Error("system+dark variant should use the dark primary colour")
	}
	if primary(themeModeSystem, theme.VariantLight) == primary(themeModeSystem, theme.VariantDark) {
		t.Error("primary colour should differ between variants")
	}
}

// TestHeadingFollowsTheme verifies headings pick up the new foreground colour when the theme changes.
func TestHeadingFollowsTheme(t *testing.T) {
	a := test.NewApp()
	h := newHeading("Solus", 30)
	w := test.NewWindow(h)
	defer w.Close()

	a.Settings().SetTheme(newSolusTheme(themeModeLight))
	h.Refresh()
	lightColor := test.WidgetRenderer(h).(*headingRenderer).label.Color

	a.Settings().SetTheme(newSolusTheme(themeModeDark))
	h.Refresh()
	darkColor := test.WidgetRenderer(h).(*headingRenderer).label.Color

	if lightColor == darkColor {
		t.Errorf("heading colour did not change between themes: %v", lightColor)
	}
}
