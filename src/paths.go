package main

import (
	"os"
	"path/filepath"
)

func findBaseDir() string {
	candidates := make([]string, 0, 4)

	if workingDir, err := os.Getwd(); err == nil {
		candidates = append(candidates, workingDir)
	}

	if executable, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executable)
		candidates = append(candidates, executableDir, filepath.Dir(executableDir))
	}

	candidates = append(candidates, "/usr/share/solus-welcome")
	for _, candidate := range candidates {
		if fileExists(filepath.Join(candidate, "assets", "logo.svg")) {
			return candidate
		}
	}

	return "."
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
