package ociinstaller

import (
	"os"
	"path/filepath"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
)

// TestDownloadImageData_InvalidLayerCount_DB tests DB downloader validation
func TestDownloadImageData_InvalidLayerCount_DB(t *testing.T) {
	downloader := newDbDownloader()

	tests := []struct {
		name    string
		layers  []ocispec.Descriptor
		wantErr bool
	}{
		{
			name:    "empty layers",
			layers:  []ocispec.Descriptor{},
			wantErr: true,
		},
		{
			name: "multiple binary layers - too many",
			layers: []ocispec.Descriptor{
				{MediaType: "application/vnd.turbot.steampipe.db.darwin-arm64.layer.v1+tar"},
				{MediaType: "application/vnd.turbot.steampipe.db.darwin-arm64.layer.v1+tar"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := downloader.GetImageData(tt.layers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImageData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Note: We got the expected error, test passes
		})
	}
}

// TestDbDownloader_EmptyConfig tests empty config creation
func TestDbDownloader_EmptyConfig(t *testing.T) {
	downloader := newDbDownloader()
	config := downloader.EmptyConfig()

	if config == nil {
		t.Error("EmptyConfig() returned nil, expected non-nil config")
	}
}

// TestDbImage_Type tests image type method
func TestDbImage_Type(t *testing.T) {
	img := &dbImage{}
	if img.Type() != ImageTypeDatabase {
		t.Errorf("Type() = %v, expected %v", img.Type(), ImageTypeDatabase)
	}
}

// TestDbDownloader_GetImageData_WithValidLayers tests successful image data extraction
func TestDbDownloader_GetImageData_WithValidLayers(t *testing.T) {
	downloader := newDbDownloader()

	// Use runtime platform to ensure test works on any OS/arch
	provider := SteampipeMediaTypeProvider{}
	mediaTypes, err := provider.MediaTypeForPlatform("db")
	if err != nil {
		t.Fatalf("Failed to get media type: %v", err)
	}

	layers := []ocispec.Descriptor{
		{
			MediaType: mediaTypes[0],
			Annotations: map[string]string{
				"org.opencontainers.image.title": "postgres-14.2",
			},
		},
		{
			MediaType: MediaTypeDbDocLayer,
			Annotations: map[string]string{
				"org.opencontainers.image.title": "README.md",
			},
		},
		{
			MediaType: MediaTypeDbLicenseLayer,
			Annotations: map[string]string{
				"org.opencontainers.image.title": "LICENSE",
			},
		},
	}

	imageData, err := downloader.GetImageData(layers)
	if err != nil {
		t.Fatalf("GetImageData() failed: %v", err)
	}

	if imageData.ArchiveDir != "postgres-14.2" {
		t.Errorf("ArchiveDir = %v, expected postgres-14.2", imageData.ArchiveDir)
	}
	if imageData.ReadmeFile != "README.md" {
		t.Errorf("ReadmeFile = %v, expected README.md", imageData.ReadmeFile)
	}
	if imageData.LicenseFile != "LICENSE" {
		t.Errorf("LicenseFile = %v, expected LICENSE", imageData.LicenseFile)
	}
}

// TestInstallDbFiles_SimpleMove tests basic installDbFiles logic
func TestInstallDbFiles_SimpleMove(t *testing.T) {
	// Create temp directories
	tempRoot := t.TempDir()
	sourceDir := filepath.Join(tempRoot, "source", "postgres-14")
	destDir := filepath.Join(tempRoot, "dest")

	// Create source with a test file
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	testFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create mock image
	mockImage := &ociinstaller.OciImage[*dbImage, *dbImageConfig]{
		Data: &dbImage{
			ArchiveDir: "postgres-14",
		},
	}

	// Call installDbFiles
	err := installDbFiles(mockImage, filepath.Join(tempRoot, "source"), destDir)
	if err != nil {
		t.Fatalf("installDbFiles failed: %v", err)
	}

	// Verify file was moved to destination
	movedFile := filepath.Join(destDir, "test.txt")
	content, err := os.ReadFile(movedFile)
	if err != nil {
		t.Errorf("Failed to read moved file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("Content mismatch: got %q, expected %q", string(content), "test content")
	}

	// Verify source is gone (MoveFolderWithinPartition should move, not copy)
	if _, err := os.Stat(sourceDir); !os.IsNotExist(err) {
		t.Error("Source directory still exists after move (expected it to be gone)")
	}
}
