package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

type DesktopInfo struct {
	Key  string
	Name string
}

// detectDesktop identifies the current desktop environment using standard Linux environment variables.
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

// architectureLabel returns a human-readable string representation of the current system architecture.
func architectureLabel() string {
	switch runtime.GOARCH {
	case "amd64", "arm64":
		return "64-bit"
	default:
		return runtime.GOARCH
	}
}

// kernelVersion reads the running Linux kernel release string from the proc filesystem.
func kernelVersion() string {
	if data, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		return strings.TrimSpace(string(data))
	}
	return ""
}

// memoryLabel reads total physical RAM from /proc/meminfo and formats it into GiB.
func memoryLabel() string {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var kb int64
				if _, err := fmt.Sscanf(fields[1], "%d", &kb); err == nil && kb > 0 {
					gib := float64(kb) / 1024 / 1024
					return fmt.Sprintf("%.1f GiB RAM", gib)
				}
			}
		}
	}
	return ""
}

// systemDetailsSummary formats architecture, kernel, and memory into a concise diagnostic string.
func systemDetailsSummary() string {
	parts := make([]string, 0, 3)
	if arch := architectureLabel(); arch != "" {
		parts = append(parts, arch)
	}
	if kernel := kernelVersion(); kernel != "" {
		parts = append(parts, "Linux "+kernel)
	}
	if mem := memoryLabel(); mem != "" {
		parts = append(parts, mem)
	}
	return strings.Join(parts, " • ")
}
