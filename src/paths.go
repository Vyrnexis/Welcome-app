package main

import (
	"os"
	"path/filepath"
)

// findBaseDir locates the application's root directory by searching common execution paths for the assets folder.
func findBaseDir() string {
	workingDir := ""
	if value, err := os.Getwd(); err == nil {
		workingDir = value
	}
	executable := ""
	if value, err := os.Executable(); err == nil {
		executable = value
	}
	for _, candidate := range baseDirCandidates(workingDir, executable) {
		if fileExists(filepath.Join(candidate, "assets", "logo.svg")) {
			return candidate
		}
	}

	return "."
}

// baseDirCandidates returns development and prefix-based asset locations in lookup order.
func baseDirCandidates(workingDir, executable string) []string {
	candidates := make([]string, 0, 5)
	if workingDir != "" {
		candidates = append(candidates, workingDir)
	}
	if executable != "" {
		executableDir := filepath.Dir(executable)
		prefixDir := filepath.Dir(executableDir)
		candidates = append(
			candidates,
			executableDir,
			prefixDir,
			filepath.Join(prefixDir, "share", "solus-welcome"),
		)
	}

	candidates = append(candidates, "/usr/share/solus-welcome")
	return candidates
}

// installationPrefix returns the prefix containing an installed share/solus-welcome directory.
func installationPrefix(baseDir string) string {
	cleanBaseDir := filepath.Clean(baseDir)
	shareDir := filepath.Dir(cleanBaseDir)
	if filepath.Base(cleanBaseDir) != "solus-welcome" || filepath.Base(shareDir) != "share" {
		return ""
	}
	return filepath.Dir(shareDir)
}

// fileExists returns a boolean indicating whether a valid regular file exists at the provided path.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
