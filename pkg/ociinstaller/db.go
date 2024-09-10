package ociinstaller

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/turbot/pipe-fittings/ociinstaller"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe/pkg/constants"
	versionfile "github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
)

// InstallDB :: Install Postgres files fom OCI image
func InstallDB(ctx context.Context, dblocation string) (string, error) {
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
