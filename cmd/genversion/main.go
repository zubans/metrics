package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/zubans/metrics/internal/version"
)

func main() {
	buildVersion := version.GetGitTag("N/A")
	buildDate := version.GetBuildDate()
	buildCommit := version.GetGitCommit()

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
		fmt.Fprintf(os.Stderr, "failed to format source: %v\n", err)
		os.Exit(1)
	}

	projectRoot := version.FindProjectRoot()
	outputPath := filepath.Join(projectRoot, "internal", "version", "version.go")
	fallbackPath := filepath.Join(projectRoot, "internal", "version", "version_fallback.go")

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		fallbackContent, err := os.ReadFile(fallbackPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to format source: %v\n", err)
			os.Exit(1)
		}
		err = os.WriteFile(outputPath, fallbackContent, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to format source: %v\n", err)
			os.Exit(1)
		}
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to format source: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(outputPath, formatted, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to format source: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Build info generated:\n")
	fmt.Printf("  Version: %s\n", buildVersion)
	fmt.Printf("  Date: %s\n", buildDate)
	fmt.Printf("  Commit: %s\n", buildCommit)
}
