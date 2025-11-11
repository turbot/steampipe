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
