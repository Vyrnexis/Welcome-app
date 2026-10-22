package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const pciDisplayClassPrefix = "0x03"

// gpuVendors names the PCI vendors worth spelling out; others are reported by their hex id.
var gpuVendors = map[string]string{
	"0x10de": "NVIDIA",
	"0x1002": "AMD",
	"0x8086": "Intel",
}

const nvidiaVendorID = "0x10de"

type gpuInfo struct {
	VendorID string
	Driver   string // empty when no kernel driver is bound
}

func (g gpuInfo) vendorName() string {
	if name, ok := gpuVendors[g.VendorID]; ok {
		return name
	}
	return g.VendorID
}

// detectGPUs lists PCI display controllers (class 0x03xxxx) found under the given sysfs root.
func detectGPUs(sysfs string) []gpuInfo {
	devices, _ := filepath.Glob(filepath.Join(sysfs, "bus", "pci", "devices", "*"))
	var gpus []gpuInfo
	for _, device := range devices {
		if !strings.HasPrefix(readTrimmed(filepath.Join(device, "class")), pciDisplayClassPrefix) {
			continue
		}
		gpu := gpuInfo{VendorID: strings.ToLower(readTrimmed(filepath.Join(device, "vendor")))}
		if target, err := os.Readlink(filepath.Join(device, "driver")); err == nil {
			gpu.Driver = filepath.Base(target)
		}
		gpus = append(gpus, gpu)
	}
	return gpus
}

// needsNvidiaDriver reports whether an NVIDIA GPU is present while the proprietary module is not loaded.
func needsNvidiaDriver(sysfs string) bool {
	for _, gpu := range detectGPUs(sysfs) {
		if gpu.VendorID == nvidiaVendorID {
			_, err := os.Stat(filepath.Join(sysfs, "module", "nvidia"))
			return err != nil
		}
	}
	return false
}

func readTrimmed(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// cpuModel returns the first "model name" entry of a /proc/cpuinfo style file.
func cpuModel(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if key, value, ok := strings.Cut(scanner.Text(), ":"); ok && strings.TrimSpace(key) == "model name" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

type reportLine struct{ label, value string }

// systemReport builds the plain-text report copied for support requests. Its labels are deliberately not
// translated: the text is pasted into English-language forums and bug trackers, where a stable format helps helpers.
func (ui *welcomeUI) systemReport() string {
	desktopName := ui.desktop.Name
	if desktopName == "" {
		desktopName = "Unknown"
	}
	if session := os.Getenv("XDG_SESSION_TYPE"); session != "" {
		desktopName += " (" + session + ")"
	}

	lines := []reportLine{
		{"OS", prettyName(ui.host.osRelease)},
		{"Desktop", desktopName},
		{"Kernel", kernelVersion()},
		{"Architecture", runtime.GOARCH},
		{"CPU", cpuModel(ui.host.cpuInfo)},
	}
	if kb := memoryKiB(); kb > 0 {
		lines = append(lines, reportLine{"Memory", fmt.Sprintf("%.1f GiB", float64(kb)/1024/1024)})
	}
	for _, gpu := range detectGPUs(ui.host.sysfs) {
		value := gpu.vendorName()
		if gpu.Driver != "" {
			value += " (driver: " + gpu.Driver + ")"
		}
		lines = append(lines, reportLine{"GPU", value})
	}

	var report strings.Builder
	for _, line := range lines {
		if line.value != "" {
			fmt.Fprintf(&report, "%s: %s\n", line.label, line.value)
		}
	}
	return report.String()
}
