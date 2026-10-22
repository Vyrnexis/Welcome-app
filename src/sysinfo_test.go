package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// addPCIDevice creates a fake sysfs PCI device; driver may be empty for an unbound device.
func addPCIDevice(t *testing.T, sysfs, addr, class, vendor, driver string) {
	t.Helper()
	device := filepath.Join(sysfs, "bus", "pci", "devices", addr)
	writeFile(t, filepath.Join(device, "class"), class+"\n")
	writeFile(t, filepath.Join(device, "vendor"), vendor+"\n")
	if driver != "" {
		target := filepath.Join(sysfs, "bus", "pci", "drivers", driver)
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(device, "driver")); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetectGPUs(t *testing.T) {
	sysfs := t.TempDir()
	addPCIDevice(t, sysfs, "0000:00:02.0", "0x030000", "0x8086", "i915")
	addPCIDevice(t, sysfs, "0000:01:00.0", "0x030200", "0x10de", "")
	addPCIDevice(t, sysfs, "0000:02:00.0", "0x030000", "0x1234", "")
	addPCIDevice(t, sysfs, "0000:00:1f.3", "0x040300", "0x8086", "snd_hda_intel") // audio, not a GPU

	gpus := detectGPUs(sysfs)
	if len(gpus) != 3 {
		t.Fatalf("detectGPUs() = %v, want 3 GPUs", gpus)
	}
	want := []struct{ vendor, driver string }{{"Intel", "i915"}, {"NVIDIA", ""}, {"0x1234", ""}}
	for i, w := range want {
		if gpus[i].vendorName() != w.vendor || gpus[i].Driver != w.driver {
			t.Errorf("gpu %d = %v, want %v", i, gpus[i], w)
		}
	}
	if got := detectGPUs(filepath.Join(sysfs, "absent")); len(got) != 0 {
		t.Errorf("detectGPUs(missing) = %v", got)
	}
}

func TestNeedsNvidiaDriver(t *testing.T) {
	sysfs := t.TempDir()
	if needsNvidiaDriver(sysfs) {
		t.Error("no GPU should not need a driver")
	}
	addPCIDevice(t, sysfs, "0000:00:02.0", "0x030000", "0x8086", "i915")
	if needsNvidiaDriver(sysfs) {
		t.Error("an Intel-only system should not need the NVIDIA driver")
	}
	addPCIDevice(t, sysfs, "0000:01:00.0", "0x030200", "0x10de", "nouveau")
	if !needsNvidiaDriver(sysfs) {
		t.Error("NVIDIA GPU without the nvidia module should need the driver")
	}
	if err := os.MkdirAll(filepath.Join(sysfs, "module", "nvidia"), 0o755); err != nil {
		t.Fatal(err)
	}
	if needsNvidiaDriver(sysfs) {
		t.Error("loaded nvidia module should hide the helper")
	}
}

func TestSystemReport(t *testing.T) {
	ui := newTestUI(t)
	ui.desktop = DesktopInfo{Key: "gnome", Name: "GNOME"}
	ui.host = hostPaths{
		osRelease: filepath.Join(t.TempDir(), "os-release"),
		cpuInfo:   filepath.Join(t.TempDir(), "cpuinfo"),
		sysfs:     t.TempDir(),
	}
	t.Setenv("XDG_SESSION_TYPE", "wayland")
	writeFile(t, ui.host.osRelease, "PRETTY_NAME=\"Solus 4.7 Endurance\"\n")
	writeFile(t, ui.host.cpuInfo, "processor\t: 0\nmodel name\t: Test CPU @ 3GHz\nmodel name\t: Second\n")
	addPCIDevice(t, ui.host.sysfs, "0000:01:00.0", "0x030000", "0x1002", "amdgpu")

	report := ui.systemReport()
	for _, want := range []string{
		"OS: Solus 4.7 Endurance\n", "Desktop: GNOME (wayland)\n", "Kernel: ", "Architecture: ",
		"CPU: Test CPU @ 3GHz\n", "Memory: ", "GPU: AMD (driver: amdgpu)\n",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("report missing %q:\n%s", want, report)
		}
	}
	if strings.Contains(report, "Second") {
		t.Errorf("report should use only the first CPU model:\n%s", report)
	}
}

// TestCopyInfoButton verifies the report reaches the clipboard and the label reverts, even on a detached button.
func TestCopyInfoButton(t *testing.T) {
	ui := newTestUI(t)
	ui.host = hostPaths{osRelease: filepath.Join(t.TempDir(), "os-release")}
	var revert func()
	var delay time.Duration
	ui.later = func(d time.Duration, f func()) { delay, revert = d, f }

	button := ui.copyInfoButton()
	original := button.Text
	button.OnTapped()

	if got := ui.app.Clipboard().Content(); !strings.Contains(got, "Architecture: ") {
		t.Errorf("clipboard = %q, want the system report", got)
	}
	if button.Text != "Copied" || delay != 2*time.Second {
		t.Errorf("after copy: text=%q delay=%v", button.Text, delay)
	}

	ui.selectPage("welcome") // rebuilding the page must not break the pending revert
	revert()
	if button.Text != original {
		t.Errorf("after revert: text=%q, want %q", button.Text, original)
	}
}
