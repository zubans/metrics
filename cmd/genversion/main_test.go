package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetGitTag(t *testing.T) {
	tests := []struct {
		name     string
		fallback string
		want     string
	}{
		{
			name:     "should return fallback when git command fails",
			fallback: "test-fallback",
			want:     "test-fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}

			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}
			defer os.Chdir(originalDir)

			got := getGitTag(tt.fallback)
			if got != tt.want {
				t.Errorf("getGitTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBuildDate(t *testing.T) {
	got := getBuildDate()

	if got == "" {
		t.Error("getBuildDate() returned empty string")
	}

	if got != "N/A" {
		if !strings.Contains(got, "_UTC") {
			t.Errorf("getBuildDate() returned invalid format: %s", got)
		}
	}
}

func TestGetGitCommit(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "should return N/A when not in git repository",
			want: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}

			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}
			defer os.Chdir(originalDir)

			got := getGitCommit()
			if got != tt.want {
				t.Errorf("getGitCommit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindProjectRoot(t *testing.T) {
	tempDir := t.TempDir()

	subDir := filepath.Join(tempDir, "subdir", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	got := findProjectRoot()
	expected := tempDir

	if got != expected {
		t.Errorf("findProjectRoot() = %v, want %v", got, expected)
	}
}

func TestFindProjectRootNoGoMod(t *testing.T) {
	tempDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Тестируем, что функция завершается с os.Exit(1)
	// Поскольку os.Exit завершает программу, мы не можем напрямую тестировать это
	// Вместо этого проверяем, что функция не возвращает результат при отсутствии go.mod
	// Это тест на интеграцию, который проверяет поведение в реальных условиях

	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to change to subdirectory: %v", err)
	}

	// Функция должна найти go.mod в родительской директории или завершиться с ошибкой
	// В данном случае она должна завершиться с ошибкой, так как go.mod нет нигде
	// Но мы не можем напрямую тестировать os.Exit, поэтому пропускаем этот тест
	t.Skip("Cannot test os.Exit behavior directly")
}

func TestMainFunction(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	internalVersionDir := filepath.Join(tempDir, "internal", "version")
	if err := os.MkdirAll(internalVersionDir, 0755); err != nil {
		t.Fatalf("Failed to create internal/version directory: %v", err)
	}

	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	fallbackPath := filepath.Join(internalVersionDir, "version_fallback.go")
	fallbackContent := `package version

var (
	BuildVersion = "N/A"
	BuildDate    = "N/A"
	BuildCommit  = "N/A"
)

func PrintBuildInfo() {
	// Fallback implementation
}
`
	if err := os.WriteFile(fallbackPath, []byte(fallbackContent), 0644); err != nil {
		t.Fatalf("Failed to create fallback file: %v", err)
	}

	buildVersion := getGitTag("N/A")
	buildDate := getBuildDate()
	buildCommit := getGitCommit()

	if buildVersion != "N/A" {
		t.Errorf("Expected buildVersion to be 'N/A', got %s", buildVersion)
	}

	if buildDate == "" {
		t.Error("Expected buildDate to be non-empty")
	}

	if buildCommit != "N/A" {
		t.Errorf("Expected buildCommit to be 'N/A', got %s", buildCommit)
	}

	projectRoot := findProjectRoot()
	if projectRoot != tempDir {
		t.Errorf("Expected projectRoot to be %s, got %s", tempDir, projectRoot)
	}
}

func createTempGitRepo(t *testing.T) string {
	tempDir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Git not available or failed to init repo: %v", err)
	}

	cmd = exec.Command("git", "tag", "v1.0.0")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to create git tag: %v", err)
	}

	return tempDir
}

func TestGetGitTagWithRealRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git not available")
	}

	tempRepo := createTempGitRepo(t)
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(tempRepo); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	got := getGitTag("fallback")
	if got != "v1.0.0" {
		t.Errorf("getGitTag() = %v, want v1.0.0", got)
	}
}
