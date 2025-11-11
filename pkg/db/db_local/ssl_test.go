package db_local

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestRemoveExpiringSelfIssuedCertificates(t *testing.T) {
	// t.Skip("Demonstrates bug #4819 - Certificate rotation without backup. Remove this skip in bug fix PR commit 1, then fix in commit 2.")

	// Setup: Create a temporary directory for testing
	tmpDir := t.TempDir()
	app_specific.InstallDir = tmpDir

	// Create the full directory structure that steampipe expects
	dataDir := filepath.Join(tmpDir, "db", constants.DatabaseVersion, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("Failed to create data directory: %v", err)
	}

	backupsDir := filepath.Join(tmpDir, "backups")
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		t.Fatalf("Failed to create backups directory: %v", err)
	}

	// Create mock server certificate and key files with "old" content
	serverCertPath := filepath.Join(dataDir, constants.ServerCert)
	serverKeyPath := filepath.Join(dataDir, constants.ServerCertKey)

	originalCertContent := []byte("OLD_CERTIFICATE_CONTENT")
	originalKeyContent := []byte("OLD_KEY_CONTENT")

	if err := os.WriteFile(serverCertPath, originalCertContent, 0644); err != nil {
		t.Fatalf("Failed to create server certificate: %v", err)
	}
	if err := os.WriteFile(serverKeyPath, originalKeyContent, 0600); err != nil {
		t.Fatalf("Failed to create server key: %v", err)
	}

	// Verify files exist before removal
	if !filehelpers.FileExists(serverCertPath) {
		t.Fatal("Server certificate should exist before removal")
	}
	if !filehelpers.FileExists(serverKeyPath) {
		t.Fatal("Server key should exist before removal")
	}

	// Simulate certificate rotation by calling removeServerCertificate
	// This should create backups before removing
	err := removeServerCertificate()
	if err != nil {
		t.Fatalf("removeServerCertificate failed: %v", err)
	}

	// Verify original certificates are removed
	if filehelpers.FileExists(serverCertPath) {
		t.Error("Server certificate should be removed")
	}
	if filehelpers.FileExists(serverKeyPath) {
		t.Error("Server key should be removed")
	}

	// TEST: Verify backups were created with timestamp
	// This is the bug - backups should exist but don't
	backupFiles, err := os.ReadDir(backupsDir)
	if err != nil {
		t.Fatalf("Failed to read backups directory: %v", err)
	}

	foundCertBackup := false
	foundKeyBackup := false
	cutoffTime := time.Now().Add(-1 * time.Minute)

	for _, file := range backupFiles {
		info, err := file.Info()
		if err != nil {
			continue
		}

		// Check if file was created recently and contains expected names
		// Backup files are named like: server.crt-2025-11-12-00-48-37
		if info.ModTime().After(cutoffTime) {
			name := file.Name()
			if filepath.Ext(name) != "" && len(name) > 4 {
				// Check if the file starts with server.crt or server.key
				if len(name) > len(constants.ServerCert) && name[:len(constants.ServerCert)] == constants.ServerCert {
					foundCertBackup = true
				}
				if len(name) > len(constants.ServerCertKey) && name[:len(constants.ServerCertKey)] == constants.ServerCertKey {
					foundKeyBackup = true
				}
			}
		}
	}

	// This assertion will FAIL demonstrating the bug
	if !foundCertBackup {
		t.Error("BUG #4819: Certificate backup was not created before removal")
	}
	if !foundKeyBackup {
		t.Error("BUG #4819: Key backup was not created before removal")
	}
}
