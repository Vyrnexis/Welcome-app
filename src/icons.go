package main

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// navIcon returns the appropriate Fyne theme icon for a given sidebar navigation item key.
func navIcon(key string) fyne.Resource {
	switch key {
	case "welcome":
		return theme.HomeIcon()
	case "customise":
		return theme.SettingsIcon()
	case "support":
		return theme.HelpIcon()
	case "contribute":
		return theme.ContentAddIcon()
	default:
		return theme.InfoIcon()
	}
}

// welcomeCardIcon maps a string action identifier to the corresponding Fyne theme icon resource.
func welcomeCardIcon(action string) fyne.Resource {
	switch action {
	case "updates":
		return theme.DownloadIcon()
	case "software":
		return theme.ComputerIcon()
	case "customise":
		return theme.ColorPaletteIcon()
	case "learn":
		return theme.DocumentIcon()
	case "settings":
		return theme.SettingsIcon()
	case "donate":
		return theme.InfoIcon()
	default:
		return theme.InfoIcon()
	}
}

// desktopIcon produces a canvas image object for the current desktop environment's logo, or a fallback icon.
func desktopIcon(baseDir string, desktop DesktopInfo) fyne.CanvasObject {
	if path := desktopIconPath(baseDir, desktop.Key); path != "" {
		icon := canvas.NewImageFromFile(path)
		icon.SetMinSize(fyne.NewSize(54, 54))
		icon.FillMode = canvas.ImageFillContain
		return icon
	}

	icon := widget.NewIcon(desktopFallbackIcon(desktop.Key))
	return minSize(container.NewCenter(icon), fyne.NewSize(54, 54))
}

// desktopFallbackIcon provides a generic Fyne resource icon when a specific desktop logo SVG cannot be found.
func desktopFallbackIcon(key string) fyne.Resource {
	switch key {
	case "kde", "gnome", "xfce", "budgie":
		return theme.ComputerIcon()
	default:
		return theme.SettingsIcon()
	}
}

// desktopIconPath searches multiple standard system locations for a specific desktop environment's icon file.
func desktopIconPath(baseDir, key string) string {
	for _, path := range desktopIconCandidates(baseDir, key) {
		if fileExists(path) {
			return path
		}
	}
	return ""
}

// desktopIconCandidates returns a list of absolute file paths to check for a desktop environment's logo.
func desktopIconCandidates(baseDir, key string) []string {
	localIcon := filepath.Join(baseDir, "assets", key+".svg")
	switch key {
	case "budgie":
		return []string{
			localIcon,
			"/usr/share/icons/hicolor/scalable/apps/budgie-desktop.svg",
			"/usr/share/icons/Papirus/64x64/apps/budgie-desktop.svg",
			"/usr/share/icons/Papirus/48x48/apps/budgie-desktop.svg",
			"/usr/share/pixmaps/budgie-desktop.svg",
			"/usr/share/pixmaps/budgie-desktop.png",
		}
	case "kde":
		return []string{
			localIcon,
			"/usr/share/icons/breeze/apps/48/start-here-kde.svg",
			"/usr/share/icons/breeze/apps/48/kde.svg",
			"/usr/share/icons/hicolor/scalable/apps/kde.svg",
			"/usr/share/icons/Papirus/64x64/places/start-here-kde.svg",
			"/usr/share/icons/Papirus/48x48/places/start-here-kde.svg",
		}
	case "gnome":
		return []string{
			localIcon,
			"/usr/share/icons/hicolor/scalable/apps/org.gnome.Shell.svg",
			"/usr/share/icons/hicolor/scalable/apps/gnome-logo-icon.svg",
			"/usr/share/icons/Adwaita/scalable/places/start-here.svg",
			"/usr/share/icons/AdwaitaLegacy/48x48/places/start-here.png",
		}
	case "xfce":
		return []string{
			localIcon,
			"/usr/share/icons/hicolor/scalable/apps/org.xfce.panel.svg",
			"/usr/share/icons/hicolor/scalable/apps/xfce4-logo.svg",
			"/usr/share/icons/Papirus/64x64/apps/xfce4-logo.svg",
			"/usr/share/icons/Papirus/48x48/apps/xfce4-logo.svg",
			"/usr/share/pixmaps/xfce4_xicon.svg",
		}
	default:
		return nil
	}
}
