package db_local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

func TestIsValidDatabaseName(t *testing.T) {
	tests := map[string]bool{
		"valid_name":  true,
		"_valid_name": true,
		"InvalidName": false,
		"123Invalid":  false,
	}

	for dbName, expectedResult := range tests {
		if actualResult := isValidDatabaseName(dbName); actualResult != expectedResult {
			t.Logf("Expected %t for %s, but for %t", expectedResult, dbName, actualResult)
			t.Fail()
		}
	}
}

func TestIsValidDatabaseName_EmptyString(t *testing.T) {
	// Test that isValidDatabaseName handles empty strings gracefully
	// An empty string should return false, not panic
	result := isValidDatabaseName("")
	if result != false {
		t.Errorf("Expected false for empty string, got %v", result)
	}
}

func TestUpdateDownloadedBinarySignatureFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	app_specific.InstallDir = filepath.Join(tempDir, ".steampipe")

	if err := updateDownloadedBinarySignature(); err != nil {
		t.Fatalf("updateDownloadedBinarySignature failed: %v", err)
	}

	sigPath := filepaths.GetDBSignatureLocation()
	info, err := os.Stat(sigPath)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", sigPath, err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected %s to have permissions 0600, got %o", sigPath, perm)
	}
}

