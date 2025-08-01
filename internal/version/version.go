// Package version содержит функциональность для работы с версией приложения.
// Включает генерацию информации о версии через go generate и функцию для вывода этой информации.

//go:generate go run ../../cmd/genversion/main.go

package version

import "fmt"

var (
	buildVersion = "N/A"
	buildDate    = "2025-07-27_12:43:44_UTC"
	buildCommit  = "3251c77666a7c95c22856b795f5ce00451204525"
)

func PrintBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
