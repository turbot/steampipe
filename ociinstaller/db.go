package ociinstaller

import (
	"context"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/constants"
	versionfile "github.com/turbot/steampipe/ociinstaller/versionfile"
)

// InstallDB :: Install Postgres files fom OCI image
func InstallDB(ctx context.Context, dblocation string) (string, error) {
	tempDir := NewTempDir(dblocation)
	defer tempDir.Delete()

	imageDownloader := NewOciDownloader()

	// Download the blobs
	image, err := imageDownloader.Download(ctx, constants.PostgresImageRef, ImageTypeDatabase, tempDir.Path)
	if err != nil {
		return "", err
	}

	// install the files
	if err = installDbFiles(image, tempDir.Path, dblocation); err != nil {
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

func installDbFiles(image *SteampipeImage, tempDir string, dest string) error {
	source := filepath.Join(tempDir, image.Database.ArchiveDir)
	return moveFolderWithinPartition(source, dest)
}
