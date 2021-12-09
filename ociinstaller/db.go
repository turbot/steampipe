package ociinstaller

import (
	"context"
	"path/filepath"
	"time"

	versionfile "github.com/turbot/steampipe/ociinstaller/versionfile"
)

// InstallDB :: Install Postgres files fom OCI image
func InstallDB(ctx context.Context, imageRef string, dest string) (string, error) {
	tempDir := NewTempDir(imageRef)
	defer tempDir.Delete()

	imageDownloader := NewOciDownloader()

	// Download the blobs
	image, err := imageDownloader.Download(ctx, imageRef, "db", tempDir.Path)
	if err != nil {
		return "", err
	}

	// install the files
	if err = extractDbFiles(image, tempDir.Path, dest); err != nil {
		return "", err
	}

	if err := updateVersionFileDB(image); err != nil {
		return string(image.OCIDescriptor.Digest), err
	}
	return string(image.OCIDescriptor.Digest), nil
}

func updateVersionFileDB(image *SteampipeImage) error {
	timeNow := versionfile.FormatTime(time.Now())
	v, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		return err
	}
	v.EmbeddedDB.Version = image.Config.Database.Version
	v.EmbeddedDB.Name = "embeddedDB"
	v.EmbeddedDB.ImageDigest = string(image.OCIDescriptor.Digest)
	v.EmbeddedDB.InstalledFrom = image.ImageRef
	v.EmbeddedDB.LastCheckedDate = timeNow
	v.EmbeddedDB.InstallDate = timeNow
	return v.Save()
}

func extractDbFiles(image *SteampipeImage, tempDir string, dest string) error {
	source := filepath.Join(tempDir, image.Database.ArchiveDir)
	return moveFolderWithinPartition(source, dest)
}
