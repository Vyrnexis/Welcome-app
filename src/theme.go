package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type solusTheme struct {
	dark bool
}

// newSolusTheme initializes a custom Fyne theme representing Solus branding with optional dark mode support.
func newSolusTheme(dark bool) fyne.Theme {
	return &solusTheme{dark: dark}
}

// Color returns a specific theme color based on the requested name, overriding the primary color with Solus blue.
func (t *solusTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	activeVariant := theme.VariantLight
	if t.dark {
		activeVariant = theme.VariantDark
	}

	if name == theme.ColorNamePrimary {
		if t.dark {
			return color.NRGBA{R: 99, G: 179, B: 237, A: 255}
		}
		return color.NRGBA{R: 49, G: 130, B: 206, A: 255}
	}

	return theme.DefaultTheme().Color(name, activeVariant)
}

// Font delegates typography requests directly to the default Fyne system theme.
func (t *solusTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon delegates all standard icon resource requests to the default Fyne system theme.
func (t *solusTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns standard dimension values from the default Fyne theme.
func (t *solusTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
