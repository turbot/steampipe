package ociinstaller

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
)

// Helper function to create a valid gzip file for testing
func createValidGzipFile(path string, content []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gzipWriter := gzip.NewWriter(f)

	_, err = gzipWriter.Write(content)
	if err != nil {
		gzipWriter.Close() // Attempt to close even on error
		return err
	}

	// Explicitly check Close() error
	if err := gzipWriter.Close(); err != nil {
		return err
	}

	return nil
}

// TestDownloadImageData_InvalidLayerCount tests validation of image layer counts
func TestDownloadImageData_InvalidLayerCount(t *testing.T) {
	// Test the validation in fdw_downloader.go:38-41 and db_downloader.go:38-41
	// These check that exactly 1 binary file is present per platform

	downloader := newFdwDownloader()

	// Test with zero layers
	emptyLayers := []ocispec.Descriptor{}
	_, err := downloader.GetImageData(emptyLayers)
	if err == nil {
		t.Error("Expected error with empty layers, got nil")
	}
	if err != nil && err.Error() != "invalid image - image should contain 1 binary file per platform, found 0" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestValidGzipFileCreation tests our helper function
func TestValidGzipFileCreation(t *testing.T) {
	tempDir := t.TempDir()
	gzipPath := filepath.Join(tempDir, "test.gz")
	expectedContent := []byte("test content for gzip")

	// Create gzip file
	if err := createValidGzipFile(gzipPath, expectedContent); err != nil {
		t.Fatalf("Failed to create gzip file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(gzipPath); os.IsNotExist(err) {
		t.Fatal("Gzip file was not created")
	}

	// Verify file size is greater than 0
	info, err := os.Stat(gzipPath)
	if err != nil {
		t.Fatalf("Failed to stat gzip file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Gzip file is empty")
	}
}

// TestMediaTypeProvider_PlatformDetection tests media type generation for different platforms
func TestMediaTypeProvider_PlatformDetection(t *testing.T) {
	provider := SteampipeMediaTypeProvider{}

	tests := []struct {
		name      string
		imageType ociinstaller.ImageType
		wantErr   bool
	}{
		{
			name:      "Database image type",
			imageType: ImageTypeDatabase,
			wantErr:   false,
		},
		{
			name:      "FDW image type",
			imageType: ImageTypeFdw,
			wantErr:   false,
		},
		{
			name:      "Plugin image type",
			imageType: ociinstaller.ImageTypePlugin,
			wantErr:   false,
		},
		{
			name:      "Assets image type",
			imageType: ImageTypeAssets,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mediaTypes, err := provider.MediaTypeForPlatform(tt.imageType)
			if (err != nil) != tt.wantErr {
				t.Errorf("MediaTypeForPlatform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(mediaTypes) == 0 && tt.imageType != ImageTypeAssets {
				t.Errorf("MediaTypeForPlatform() returned empty media types for %s", tt.imageType)
			}
		})
	}
}

// TestInstallFdwFiles_CorruptGzipFile_BugDocumentation documents bug #4753
// This test documents the critical bug where the existing FDW binary was deleted
// before verifying that the new binary could be successfully extracted.
//
// Bug Scenario (BEFORE FIX):
// 1. User has working FDW v1.0 installed
// 2. Upgrade to v2.0 begins
// 3. os.Remove() deletes the v1.0 binary (line 70 in fdw.go)
// 4. Ungzip() attempts to extract v2.0 binary (line 72)
// 5. If ungzip fails (corrupt download, disk full, etc.):
//    - Old v1.0 binary is GONE (deleted in step 3)
//    - New v2.0 binary FAILED to install (step 4)
//    - System is now BROKEN with no FDW at all
//
// This test simulates the old buggy behavior for documentation purposes.
// It is skipped because it will always fail (it simulates the bug itself).
// The fix ensures this scenario can never happen in the actual code.
func TestInstallFdwFiles_CorruptGzipFile_BugDocumentation(t *testing.T) {
	t.Skip("Documentation test - simulates the bug that existed before fix #4753")

	// Setup: Create temp directories to simulate FDW installation directories
	tempInstallDir := t.TempDir()
	tempSourceDir := t.TempDir()

	// Create a valid "existing" FDW binary (v1.0)
	existingBinaryPath := filepath.Join(tempInstallDir, "steampipe-postgres-fdw.so")
	existingBinaryContent := []byte("existing FDW v1.0 binary")
	if err := os.WriteFile(existingBinaryPath, existingBinaryContent, 0755); err != nil {
		t.Fatalf("Failed to create existing FDW binary: %v", err)
	}

	// Create a CORRUPT gzip file (not a valid gzip) that will fail to ungzip
	corruptGzipPath := filepath.Join(tempSourceDir, "steampipe-postgres-fdw.so.gz")
	corruptGzipContent := []byte("this is not a valid gzip file, ungzip will fail")
	if err := os.WriteFile(corruptGzipPath, corruptGzipContent, 0644); err != nil {
		t.Fatalf("Failed to create corrupt gzip file: %v", err)
	}

	// Simulate the OLD BUGGY behavior from installFdwFiles() (before fix):
	// 1. Remove the old binary first
	// 2. Then try to ungzip (which will fail with our corrupt file)
	os.Remove(existingBinaryPath)
	_, ungzipErr := ociinstaller.Ungzip(corruptGzipPath, tempInstallDir)

	// Verify ungzip failed (confirms test setup)
	if ungzipErr == nil {
		t.Fatal("Expected ungzip to fail with corrupt file, but it succeeded")
	}

	// CRITICAL ASSERTION: After a failed ungzip, the old binary should still exist
	// But with the buggy code, it's gone!
	_, statErr := os.Stat(existingBinaryPath)
	if os.IsNotExist(statErr) {
		// This demonstrates the bug: The old binary was deleted BEFORE verifying
		// that the new binary could be successfully extracted.
		t.Errorf("CRITICAL BUG: Old FDW binary was deleted before new binary extraction succeeded. System left in broken state with no FDW binary.")
	}
}

// TestInstallFdwFiles_PartialInstall_BugDocumentation demonstrates the non-atomic installation bug
// where failure during installation can leave the system in an inconsistent state with a mix of
// old and new files. This test simulates a failure during the control file move operation.
//
// Bug: installFdwFiles performs three sequential file operations without atomicity:
// 1. Ungzip binary to destination
// 2. Move control file
// 3. Move SQL file
//
// If step 2 or 3 fails, the binary is already in place with the new version but control/SQL
// files are old version or missing, causing FDW to fail to load or behave unpredictably.
//
// Expected behavior: Installation should be atomic - either all files are updated or none are.
func TestInstallFdwFiles_PartialInstall_BugDocumentation(t *testing.T) {
	// Create temp directories simulating installation source and destination
	tempRoot := t.TempDir()
	sourceDir := filepath.Join(tempRoot, "source")
	binDestDir := filepath.Join(tempRoot, "fdw_bin")
	controlDestDir := filepath.Join(tempRoot, "fdw_control")

	// Create source and destination directories
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	if err := os.MkdirAll(binDestDir, 0755); err != nil {
		t.Fatalf("Failed to create bin dest dir: %v", err)
	}
	if err := os.MkdirAll(controlDestDir, 0755); err != nil {
		t.Fatalf("Failed to create control dest dir: %v", err)
	}

	// Create v2.0 binary file (uncompressed for simplicity)
	binarySourcePath := filepath.Join(sourceDir, "steampipe_postgres_fdw.so")
	if err := os.WriteFile(binarySourcePath, []byte("v2.0 binary content"), 0644); err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create control and SQL files (simulating v2.0)
	controlSourcePath := filepath.Join(sourceDir, "steampipe_postgres_fdw.control")
	if err := os.WriteFile(controlSourcePath, []byte("v2.0 control"), 0644); err != nil {
		t.Fatalf("Failed to create control file: %v", err)
	}

	sqlSourcePath := filepath.Join(sourceDir, "steampipe_postgres_fdw--1.0.sql")
	if err := os.WriteFile(sqlSourcePath, []byte("v2.0 sql"), 0644); err != nil {
		t.Fatalf("Failed to create SQL file: %v", err)
	}

	// Simulate existing v1.0 installation
	binPath := filepath.Join(binDestDir, "steampipe_postgres_fdw.so")
	if err := os.WriteFile(binPath, []byte("v1.0 binary"), 0644); err != nil {
		t.Fatalf("Failed to create old binary: %v", err)
	}

	oldControlPath := filepath.Join(controlDestDir, "steampipe_postgres_fdw.control")
	if err := os.WriteFile(oldControlPath, []byte("v1.0 control"), 0644); err != nil {
		t.Fatalf("Failed to create old control: %v", err)
	}

	oldSqlPath := filepath.Join(controlDestDir, "steampipe_postgres_fdw--1.0.sql")
	if err := os.WriteFile(oldSqlPath, []byte("v1.0 sql"), 0644); err != nil {
		t.Fatalf("Failed to create old SQL: %v", err)
	}

	// TEST: Simulate installation failure after binary copy but before control file move
	// This demonstrates the non-atomic nature of the current implementation

	// Step 1: Copy binary to destination (this succeeds - simulating Ungzip)
	os.Remove(binPath) // Remove old binary as the code does
	// Copy file manually
	srcFile, err := os.Open(binarySourcePath)
	if err != nil {
		t.Fatalf("Failed to open source binary: %v", err)
	}
	dstFile, err := os.Create(binPath)
	if err != nil {
		srcFile.Close()
		t.Fatalf("Failed to create dest binary: %v", err)
	}
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		srcFile.Close()
		dstFile.Close()
		t.Fatalf("Binary copy failed: %v", err)
	}
	srcFile.Close()
	dstFile.Close()

	// Verify new binary is in place
	newBinContent, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("Failed to read new binary: %v", err)
	}
	if string(newBinContent) != "v2.0 binary content" {
		t.Errorf("Binary not updated correctly")
	}

	// Step 2: Simulate control file move failure (by making destination read-only)
	// This is where the bug manifests - binary is v2.0 but other files are still v1.0
	if err := os.Chmod(controlDestDir, 0555); err != nil {
		t.Fatalf("Failed to make control dir read-only: %v", err)
	}
	defer os.Chmod(controlDestDir, 0755) // Restore permissions for cleanup

	// Attempt to move control file (this will fail)
	controlDestPath := filepath.Join(controlDestDir, "steampipe_postgres_fdw.control")
	err = ociinstaller.MoveFileWithinPartition(controlSourcePath, controlDestPath)

	// BUG DEMONSTRATION: At this point we have an inconsistent state
	// - Binary is v2.0 (new)
	// - Control file is v1.0 (old)
	// - SQL file is v1.0 (old)
	// This is the bug: installation is not atomic!

	if err == nil {
		t.Error("Expected control file move to fail, but it succeeded")
	}

	// Verify inconsistent state: new binary with old control/SQL files
	binContent, _ := os.ReadFile(binPath)
	controlContent, _ := os.ReadFile(oldControlPath)
	sqlContent, _ := os.ReadFile(oldSqlPath)

	if string(binContent) != "v2.0 binary content" {
		t.Errorf("Binary should be v2.0, got: %s", string(binContent))
	}

	if string(controlContent) != "v1.0 control" {
		t.Errorf("Control should still be v1.0, got: %s", string(controlContent))
	}

	if string(sqlContent) != "v1.0 sql" {
		t.Errorf("SQL should still be v1.0, got: %s", string(sqlContent))
	}

	// This inconsistent state (v2.0 binary + v1.0 control/SQL) is the bug!
	// The FDW will fail to load or behave unpredictably
	t.Logf("BUG CONFIRMED: System left in inconsistent state - binary v2.0, control v1.0, SQL v1.0")
	t.Error("Installation is not atomic - partial failure leaves system in inconsistent state")
}
