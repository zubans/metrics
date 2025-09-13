package version

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetGitTag returns the latest git tag or fallback if command fails
func GetGitTag(fallback string) string {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return fallback
	}
	return strings.TrimSpace(string(output))
}

// GetBuildDate returns current UTC date in format YYYY-MM-DD_HH:MM:SS_UTC
func GetBuildDate() string {
	cmd := exec.Command("date", "-u", "+%Y-%m-%d_%H:%M:%S_UTC")
	output, err := cmd.Output()
	if err != nil {
		return "N/A"
	}
	return strings.TrimSpace(string(output))
}

// GetGitCommit returns current git commit hash
func GetGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "N/A"
	}
	return strings.TrimSpace(string(output))
}

// FindProjectRoot finds the project root directory by looking for go.mod
func FindProjectRoot() string {
	currentDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			return ""
		}
		currentDir = parent
	}
}
