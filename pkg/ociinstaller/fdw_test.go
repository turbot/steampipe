package ociinstaller

import (
	"compress/gzip"
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

