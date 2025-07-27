// Package version содержит функциональность для работы с версией приложения.
// Включает генерацию информации о версии через go generate и функцию для вывода этой информации.

//go:generate go run ../../cmd/genversion/main.go

package version

import "fmt"

var (
	BuildVersion = "N/A"
	BuildDate    = "2025-07-27_12:43:44_UTC"
	BuildCommit  = "3251c77666a7c95c22856b795f5ce00451204525"
)

func PrintBuildInfo() {
	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Build commit: %s\n", BuildCommit)
}
