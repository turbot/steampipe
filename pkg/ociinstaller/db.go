package ociinstaller

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	versionfile "github.com/turbot/steampipe/v2/pkg/ociinstaller/versionfile"
)

// InstallDB :: Install Postgres files fom OCI image
func InstallDB(ctx context.Context, dblocation string) (string, error) {
	// Check available disk space BEFORE starting installation
	// This prevents partial installations that can leave the system in a broken state
	if err := validateDiskSpace(dblocation, constants.PostgresImageRef); err != nil {
		return "", err
	}

	tempDir := ociinstaller.NewTempDir(dblocation)
	defer func() {
		if err := tempDir.Delete(); err != nil {
			log.Printf("[TRACE] Failed to delete temp dir '%s' after installing db files: %s", tempDir, err)
		}
	}()

	imageDownloader := newDbDownloader()

	// Download the blobs
	image, err := imageDownloader.Download(ctx, ociinstaller.NewImageRef(constants.PostgresImageRef), ImageTypeDatabase, tempDir.Path)
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

func updateVersionFileDB(image *ociinstaller.OciImage[*dbImage, *dbImageConfig]) error {
	timeNow := utils.FormatTime(time.Now())
	v, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		return err
	}
	v.EmbeddedDB.Version = image.Config.Database.Version
	v.EmbeddedDB.Name = "embeddedDB"
	v.EmbeddedDB.ImageDigest = string(image.OCIDescriptor.Digest)
	v.EmbeddedDB.InstalledFrom = image.ImageRef.RequestedRef
	v.EmbeddedDB.LastCheckedDate = timeNow
	v.EmbeddedDB.InstallDate = timeNow
	return v.Save()
}

func installDbFiles(image *ociinstaller.OciImage[*dbImage, *dbImageConfig], tempDir string, dest string) error {
	source := filepath.Join(tempDir, image.Data.ArchiveDir)
	return ociinstaller.MoveFolderWithinPartition(source, dest)
}
