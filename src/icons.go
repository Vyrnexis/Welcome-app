package main

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

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

func welcomeCardIcon(action welcomeAction) fyne.Resource {
	switch action {
	case actionUpdates:
		return theme.DownloadIcon()
	case actionSoftware:
		return theme.ComputerIcon()
	case actionCustomise:
		return theme.ColorPaletteIcon()
	case actionLearn:
		return theme.DocumentIcon()
	case actionSettings:
		return theme.SettingsIcon()
	case actionDonate:
		return theme.InfoIcon()
	default:
		return theme.InfoIcon()
	}
}

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

func desktopFallbackIcon(key string) fyne.Resource {
	switch key {
	case "kde", "gnome", "xfce", "budgie":
		return theme.ComputerIcon()
	default:
		return theme.SettingsIcon()
	}
}

func desktopIconPath(baseDir, key string) string {
	for _, path := range desktopIconCandidates(baseDir, key) {
		if fileExists(path) {
			return path
		}
	}
	return ""
}

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
