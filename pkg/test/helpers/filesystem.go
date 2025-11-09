package helpers

import (
	"os"
	"path/filepath"
	"testing"
)

// CreateTempDir creates a temporary directory for testing and registers cleanup
func CreateTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "steampipe-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// WriteTestFile writes content to a file in a temp directory
func WriteTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)

	// Create parent directories if needed
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	return path
}

// CreateTestConfigFile creates a test .spc config file
func CreateTestConfigFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	return WriteTestFile(t, dir, name+".spc", content)
}

// CreateTestDir creates a directory in a test directory
func CreateTestDir(t *testing.T, baseDir, dirname string) string {
	t.Helper()
	path := filepath.Join(baseDir, dirname)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	return path
}

// ReadTestFile reads a file's content for verification in tests
func ReadTestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	return string(content)
}

// FileExists checks if a file exists (useful for assertions)
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists (useful for assertions)
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
