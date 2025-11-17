package ociinstaller

import (
	"context"
	"fmt"
	"log"
	"os"
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
		return "", err
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

	// For atomic installation, we use a staging approach:
	// 1. Create a staging directory next to the destination
	// 2. Move all files to staging first (this validates all operations can succeed)
	// 3. Atomically rename staging directory to destination
	//
	// This ensures either all files are updated or none are, avoiding inconsistent states

	// Create staging directory next to destination for atomic swap
	stagingDest := dest + ".staging"
	backupDest := dest + ".backup"

	// Clean up any previous failed installation attempts
	// This handles cases where the process was killed during installation
	os.RemoveAll(stagingDest)
	os.RemoveAll(backupDest)

	// Move source to staging location
	if err := ociinstaller.MoveFolderWithinPartition(source, stagingDest); err != nil {
		return err
	}

	// Now atomically swap: rename old dest as backup, rename staging to dest
	// If destination exists, rename it to backup location
	destExists := false
	if _, err := os.Stat(dest); err == nil {
		destExists = true
		// Attempt atomic rename of old installation to backup
		if err := os.Rename(dest, backupDest); err != nil {
			// Failed to backup old installation - abort and restore staging
			// Move staging back to source if possible
			os.RemoveAll(stagingDest)
			return fmt.Errorf("could not backup existing installation: %s", err.Error())
		}
	}

	// Atomically move staging to final destination
	if err := os.Rename(stagingDest, dest); err != nil {
		// Failed to move staging to destination
		// Try to restore backup if it exists
		if destExists {
			os.Rename(backupDest, dest)
		}
		return fmt.Errorf("could not install database files: %s", err.Error())
	}

	// Success - clean up backup
	if destExists {
		os.RemoveAll(backupDest)
	}

	return nil
}
