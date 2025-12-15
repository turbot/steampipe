package ociinstaller

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	putils "github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	"github.com/turbot/steampipe/v2/pkg/ociinstaller/versionfile"
)

// InstallFdw installs the Steampipe Postgres foreign data wrapper from an OCI image
func InstallFdw(ctx context.Context, dbLocation string) (string, error) {
	// Check available disk space BEFORE starting installation
	// This prevents partial installations that can leave the system in a broken state
	if err := validateDiskSpace(dbLocation, constants.FdwImageRef); err != nil {
		return "", err
	}

	tempDir := ociinstaller.NewTempDir(dbLocation)
	defer func() {
		if err := tempDir.Delete(); err != nil {
			log.Printf("[TRACE] Failed to delete temp dir '%s' after installing fdw: %s", tempDir, err)
		}
	}()

	imageDownloader := newFdwDownloader()

	// download the blobs.
	image, err := imageDownloader.Download(ctx, ociinstaller.NewImageRef(constants.FdwImageRef), ImageTypeFdw, tempDir.Path)
	if err != nil {
		return "", err
	}

	// install the files
	if err = installFdwFiles(image, tempDir.Path); err != nil {
		return "", err
	}

	if err := updateVersionFileFdw(image); err != nil {
		return "", err
	}

	return string(image.OCIDescriptor.Digest), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Sync to ensure data is written
	return destFile.Sync()
}

func updateVersionFileFdw(image *ociinstaller.OciImage[*fdwImage, *FdwImageConfig]) error {
	timeNow := putils.FormatTime(time.Now())
	v, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		return err
	}
	v.FdwExtension.Version = image.Config.Fdw.Version
	v.FdwExtension.Name = "fdwExtension"
	v.FdwExtension.ImageDigest = string(image.OCIDescriptor.Digest)
	v.FdwExtension.InstalledFrom = image.ImageRef.RequestedRef
	v.FdwExtension.LastCheckedDate = timeNow
	v.FdwExtension.InstallDate = timeNow
	return v.Save()
}

func installFdwFiles(image *ociinstaller.OciImage[*fdwImage, *FdwImageConfig], tempdir string) error {
	// Create staging directory for atomic installation
	// All files will be prepared in staging first, then moved atomically to their final locations
	stagingDir := filepath.Join(tempdir, "staging")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return fmt.Errorf("could not create staging directory: %s", err.Error())
	}

	// Determine final destination paths
	fdwBinDir := filepaths.GetFDWBinaryDir()
	fdwControlDir := filepaths.GetFDWSQLAndControlDir()
	fdwSQLDir := filepaths.GetFDWSQLAndControlDir()

	fdwBinFileSourcePath := filepath.Join(tempdir, image.Data.BinaryFile)
	controlFileSourcePath := filepath.Join(tempdir, image.Data.ControlFile)
	sqlFileSourcePath := filepath.Join(tempdir, image.Data.SqlFile)

	// Stage 1: Extract and stage all files to staging directory
	// If any operation fails here, no destination files have been touched yet

	// Stage binary: ungzip to staging directory
	stagingBinDir := filepath.Join(stagingDir, "bin")
	if err := os.MkdirAll(stagingBinDir, 0755); err != nil {
		return fmt.Errorf("could not create staging bin directory: %s", err.Error())
	}

	stagedBinaryPath, err := ociinstaller.Ungzip(fdwBinFileSourcePath, stagingBinDir)
	if err != nil {
		return fmt.Errorf("could not unzip %s to staging: %s", fdwBinFileSourcePath, err.Error())
	}

	// Stage control file: copy to staging
	stagingControlPath := filepath.Join(stagingDir, image.Data.ControlFile)
	if err := copyFile(controlFileSourcePath, stagingControlPath); err != nil {
		return fmt.Errorf("could not stage control file %s: %s", controlFileSourcePath, err.Error())
	}

	// Stage SQL file: copy to staging
	stagingSQLPath := filepath.Join(stagingDir, image.Data.SqlFile)
	if err := copyFile(sqlFileSourcePath, stagingSQLPath); err != nil {
		return fmt.Errorf("could not stage SQL file %s: %s", sqlFileSourcePath, err.Error())
	}

	// Stage 2: All files staged successfully - now atomically move them to final destinations
	// NOTE: for Mac M1 machines, if the fdw binary is updated in place without deleting the existing file,
	// the updated fdw may crash on execution - for an undetermined reason
	// To avoid this AND prevent leaving the system without a binary if the move fails,
	// we move to a temp location first, then delete old, then rename to final location
	fdwBinFileDestPath := filepath.Join(fdwBinDir, constants.FdwBinaryFileName)
	tempBinaryPath := fdwBinFileDestPath + ".tmp"

	// Move staged binary to temp location first (verifies the move works)
	if err := ociinstaller.MoveFileWithinPartition(stagedBinaryPath, tempBinaryPath); err != nil {
		return fmt.Errorf("could not move binary from staging to temp location: %s", err.Error())
	}

	// Now that we know the new binary is ready, remove the old one
	os.Remove(fdwBinFileDestPath)

	// Finally, atomically rename temp to final location
	if err := os.Rename(tempBinaryPath, fdwBinFileDestPath); err != nil {
		return fmt.Errorf("could not install binary to %s: %s", fdwBinDir, err.Error())
	}

	// Move staged control file to destination
	controlFileDestPath := filepath.Join(fdwControlDir, image.Data.ControlFile)
	if err := ociinstaller.MoveFileWithinPartition(stagingControlPath, controlFileDestPath); err != nil {
		// Binary was already moved - try to rollback by removing it
		os.Remove(fdwBinFileDestPath)
		return fmt.Errorf("could not install control file from staging to %s: %s", fdwControlDir, err.Error())
	}

	// Move staged SQL file to destination
	sqlFileDestPath := filepath.Join(fdwSQLDir, image.Data.SqlFile)
	if err := ociinstaller.MoveFileWithinPartition(stagingSQLPath, sqlFileDestPath); err != nil {
		// Binary and control were already moved - try to rollback
		os.Remove(fdwBinFileDestPath)
		os.Remove(controlFileDestPath)
		return fmt.Errorf("could not install SQL file from staging to %s: %s", fdwSQLDir, err.Error())
	}

	return nil
}
