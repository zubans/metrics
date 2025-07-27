package main

import (
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	buildVersion := getGitTag()
	buildDate := getBuildDate()
	buildCommit := getGitCommit()

	var sb strings.Builder
	sb.WriteString(`// Package version содержит функциональность для работы с версией приложения.
// Включает генерацию информации о версии через go generate и функцию для вывода этой информации.

//go:generate go run ../../cmd/genversion/main.go

package version

import "fmt"

// Fallback значения (используются по умолчанию)
var (
	BuildVersion = `)
	sb.WriteString(fmt.Sprintf("%q", buildVersion))
	sb.WriteString(`
	BuildDate    = `)
	sb.WriteString(fmt.Sprintf("%q", buildDate))
	sb.WriteString(`
	BuildCommit  = `)
	sb.WriteString(fmt.Sprintf("%q", buildCommit))
	sb.WriteString(`
)

func PrintBuildInfo() {
	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Build commit: %s\n", BuildCommit)
}
`)

	generated := []byte(sb.String())
	formatted, err := format.Source(generated)
	if err != nil {
		panic(err)
	}

	projectRoot := findProjectRoot()
	outputPath := filepath.Join(projectRoot, "internal", "version", "version.go")
	fallbackPath := filepath.Join(projectRoot, "internal", "version", "version_fallback.go")

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		fallbackContent, err := os.ReadFile(fallbackPath)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(outputPath, fallbackContent, 0644)
		if err != nil {
			panic(err)
		}
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	err = os.WriteFile(outputPath, formatted, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Build info generated:\n")
	fmt.Printf("  Version: %s\n", buildVersion)
	fmt.Printf("  Date: %s\n", buildDate)
	fmt.Printf("  Commit: %s\n", buildCommit)
}

func findProjectRoot() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			panic("go.mod not found in any parent directory")
		}
		currentDir = parent
	}
}

func getGitTag() string {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "N/A"
	}
	return strings.TrimSpace(string(output))
}

func getBuildDate() string {
	cmd := exec.Command("date", "-u", "+%Y-%m-%d_%H:%M:%S_UTC")
	output, err := cmd.Output()
	if err != nil {
		return "N/A"
	}
	return strings.TrimSpace(string(output))
}

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "N/A"
	}
	return strings.TrimSpace(string(output))
}
