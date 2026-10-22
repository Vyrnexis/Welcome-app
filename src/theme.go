package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type themeMode string

const (
	themeModeSystem themeMode = "system"
	themeModeLight  themeMode = "light"
	themeModeDark   themeMode = "dark"

	themeModePreferenceKey = "themeMode"
	legacyDarkThemeKey     = "darkTheme"
)

// themeModes is the order shown in the theme selector.
var themeModes = []themeMode{themeModeSystem, themeModeLight, themeModeDark}

// hasLegacyDarkTheme reports whether the old bool key is stored; Preferences has no direct presence check.
func hasLegacyDarkTheme(prefs fyne.Preferences) bool {
	return prefs.BoolWithFallback(legacyDarkThemeKey, false) == prefs.BoolWithFallback(legacyDarkThemeKey, true)
}

// loadThemeMode reads the saved mode, migrating the legacy boolean "darkTheme" preference once.
func loadThemeMode(prefs fyne.Preferences) themeMode {
	if hasLegacyDarkTheme(prefs) {
		defer prefs.RemoveValue(legacyDarkThemeKey)
		if prefs.StringWithFallback(themeModePreferenceKey, "") == "" {
			mode := themeModeSystem
			if prefs.Bool(legacyDarkThemeKey) {
				mode = themeModeDark
			}
			prefs.SetString(themeModePreferenceKey, string(mode))
			return mode
		}
	}

	switch mode := themeMode(prefs.StringWithFallback(themeModePreferenceKey, "")); mode {
	case themeModeLight, themeModeDark:
		return mode
	default:
		return themeModeSystem
	}
}

type solusTheme struct {
	mode themeMode
}

func newSolusTheme(mode themeMode) fyne.Theme {
	return &solusTheme{mode: mode}
}

// Color follows the variant Fyne passes (the desktop preference) unless a mode forces light or dark.
func (t *solusTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch t.mode {
	case themeModeLight:
		variant = theme.VariantLight
	case themeModeDark:
		variant = theme.VariantDark
	}

	switch name {
	case theme.ColorNamePrimary:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 99, G: 179, B: 237, A: 255}
		}
		return color.NRGBA{R: 49, G: 130, B: 206, A: 255}
	case theme.ColorNameForegroundOnPrimary:
		// White on the light dark-mode blue is about 2:1; near-black keeps button labels readable.
		if variant == theme.VariantDark {
			return color.NRGBA{R: 26, G: 32, B: 44, A: 255}
		}
		return color.White
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (t *solusTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *solusTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *solusTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// heading is a bold title that re-reads the foreground colour on every refresh, so it follows theme changes.
type heading struct {
	widget.BaseWidget
	text string
	size float32
}

func newHeading(text string, size float32) *heading {
	h := &heading{text: text, size: size}
	h.ExtendBaseWidget(h)
	return h
}

func (h *heading) CreateRenderer() fyne.WidgetRenderer {
	r := &headingRenderer{heading: h, label: canvas.NewText(h.text, nil)}
	r.label.TextStyle = fyne.TextStyle{Bold: true}
	r.Refresh()
	return r
}

type headingRenderer struct {
	heading *heading
	label   *canvas.Text
}

func (r *headingRenderer) Layout(size fyne.Size) {
	r.label.Move(fyne.NewPos(0, 0))
	r.label.Resize(size)
}

func (r *headingRenderer) MinSize() fyne.Size {
	return r.label.MinSize()
}

func (r *headingRenderer) Refresh() {
	r.label.Text = r.heading.text
	r.label.TextSize = r.heading.size
	r.label.Color = theme.Color(theme.ColorNameForeground)
	r.label.Refresh()
}

func (r *headingRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.label}
}

func (r *headingRenderer) Destroy() {}
