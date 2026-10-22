package main

import (
	"bufio"
	"os"
	"strings"
)

// hostPaths locates the system files the app reads, so tests can point them at fixtures.
type hostPaths struct {
	osRelease string
	cpuInfo   string
	sysfs     string
}

var defaultHostPaths = hostPaths{
	osRelease: "/etc/os-release",
	cpuInfo:   "/proc/cpuinfo",
	sysfs:     "/sys",
}

// parseOSRelease reads os-release style KEY=value lines, ignoring comments and unquoting values.
// A missing or unreadable file yields an empty map.
func parseOSRelease(path string) map[string]string {
	values := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		return values
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		values[strings.TrimSpace(key)] = value
	}
	return values
}

// prettyName returns the distribution's PRETTY_NAME, or "" when it is not available.
func prettyName(path string) string {
	return parseOSRelease(path)["PRETTY_NAME"]
}
