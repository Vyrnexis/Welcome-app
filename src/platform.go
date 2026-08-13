package main

import (
	"os"
	"runtime"
	"strings"
)

type DesktopInfo struct {
	Key  string
	Name string
}

func detectDesktop() DesktopInfo {
	values := []string{
		os.Getenv("XDG_CURRENT_DESKTOP"),
		os.Getenv("DESKTOP_SESSION"),
		os.Getenv("GDMSESSION"),
	}

	for _, value := range values {
		desktop := strings.ToLower(value)
		switch {
		case strings.Contains(desktop, "kde"), strings.Contains(desktop, "plasma"):
			return DesktopInfo{Key: "kde", Name: "KDE Plasma"}
		case strings.Contains(desktop, "budgie"):
			return DesktopInfo{Key: "budgie", Name: "Budgie"}
		case strings.Contains(desktop, "gnome"):
			return DesktopInfo{Key: "gnome", Name: "GNOME"}
		case strings.Contains(desktop, "xfce"), strings.Contains(desktop, "xfce4"):
			return DesktopInfo{Key: "xfce", Name: "Xfce"}
		}
	}

	return DesktopInfo{Key: "unknown", Name: "Unknown desktop"}
}

func architectureLabel() string {
	switch runtime.GOARCH {
	case "amd64", "arm64":
		return "64-bit"
	default:
		return runtime.GOARCH
	}
}
