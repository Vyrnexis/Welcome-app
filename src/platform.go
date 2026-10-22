package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"fyne.io/fyne/v2/lang"
)

type DesktopInfo struct {
	Key  string
	Name string
}

type desktopSpec struct {
	key     string
	name    string
	markers []string
}

// desktops lists supported desktops in match priority order: Budgie sessions also advertise GNOME.
var desktops = []desktopSpec{
	{"kde", "KDE Plasma", []string{"kde", "plasma"}},
	{"budgie", "Budgie", []string{"budgie"}},
	{"gnome", "GNOME", []string{"gnome"}},
	{"xfce", "Xfce", []string{"xfce"}},
}

// knownDesktop reports whether key identifies a supported desktop.
func knownDesktop(key string) bool {
	for _, spec := range desktops {
		if spec.key == key {
			return true
		}
	}
	return false
}

// detectDesktop identifies the current desktop environment using standard Linux environment variables.
func detectDesktop() DesktopInfo {
	values := []string{
		os.Getenv("XDG_CURRENT_DESKTOP"),
		os.Getenv("DESKTOP_SESSION"),
		os.Getenv("GDMSESSION"),
	}

	for _, value := range values {
		value = strings.ToLower(value)
		for _, spec := range desktops {
			for _, marker := range spec.markers {
				if strings.Contains(value, marker) {
					return DesktopInfo{Key: spec.key, Name: spec.name}
				}
			}
		}
	}

	return DesktopInfo{Key: "unknown"}
}

// architectureLabel returns a human-readable string representation of the current system architecture.
func architectureLabel(ui UIStrings) string {
	switch runtime.GOARCH {
	case "amd64", "arm64":
		return lang.L(ui.Bits64)
	default:
		return runtime.GOARCH
	}
}

func kernelVersion() string {
	if data, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		return strings.TrimSpace(string(data))
	}
	return ""
}

// memoryKiB reads total physical RAM from /proc/meminfo, or 0 when it cannot be determined.
func memoryKiB() int64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var kb int64
				if _, err := fmt.Sscanf(fields[1], "%d", &kb); err == nil && kb > 0 {
					return kb
				}
			}
		}
	}
	return 0
}

// memoryLabel formats total physical RAM into GiB.
func memoryLabel(ui UIStrings) string {
	if kb := memoryKiB(); kb > 0 {
		return fmt.Sprintf(lang.L(ui.MemoryFormat), float64(kb)/1024/1024)
	}
	return ""
}

// systemDetailsSummary formats architecture, kernel, and memory into a concise diagnostic string.
func systemDetailsSummary(ui UIStrings) string {
	parts := make([]string, 0, 3)
	if arch := architectureLabel(ui); arch != "" {
		parts = append(parts, arch)
	}
	if kernel := kernelVersion(); kernel != "" {
		parts = append(parts, "Linux "+kernel)
	}
	if mem := memoryLabel(ui); mem != "" {
		parts = append(parts, mem)
	}
	return strings.Join(parts, " • ")
}
