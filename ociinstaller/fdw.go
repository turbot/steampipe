package ociinstaller

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/constants"
	versionfile "github.com/turbot/steampipe/ociinstaller/versionfile"
	"github.com/turbot/steampipe/utils"
)

// InstallFdw installs the Steampipe Postgres foreign data wrapper from an OCI image
func InstallFdw(ctx context.Context, dbLocation string) (string, error) {
	tempDir := NewTempDir(dbLocation)
	defer func() {
		if err := tempDir.Delete(); err != nil {
			log.Printf("[TRACE] Failed to delete temp dir '%s' after installing fdw: %s", tempDir, err)
		}
	}()

	imageDownloader := NewOciDownloader()

	// download the blobs.
	image, err := imageDownloader.Download(ctx, NewSteampipeImageRef(constants.FdwImageRef), ImageTypeFdw, tempDir.Path)
	if err != nil {
		return "", err
	}

	// install the files
	if err = installFdwFiles(image, tempDir.Path, dbLocation); err != nil {
		return "", err
	}

	if err := updateVersionFileFdw(image); err != nil {
		return string(image.OCIDescriptor.Digest), err
	}

	return string(image.OCIDescriptor.Digest), nil
}

func updateVersionFileFdw(image *SteampipeImage) error {
	timeNow := versionfile.FormatTime(time.Now())
	v, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		return err
	}
	v.FdwExtension.Version = image.Config.Fdw.Version
	v.FdwExtension.Name = "fdwExtension"
	v.FdwExtension.ImageDigest = string(image.OCIDescriptor.Digest)
	v.FdwExtension.InstalledFrom = image.ImageRef.requestedRef
	v.FdwExtension.LastCheckedDate = timeNow
	v.FdwExtension.InstallDate = timeNow
	return v.Save()
}

func installFdwFiles(image *SteampipeImage, tempdir string, dest string) error {
	hubBinPath := filepath.Join(dest, "lib", "postgresql")
	hubControlPath := filepath.Join(dest, "share", "postgresql", "extension")
	hubSQLPath := filepath.Join(dest, "share", "postgresql", "extension")
	fileName := image.Fdw.BinaryFile
	sourcePath := filepath.Join(tempdir, fileName)

	isM1, err := utils.IsMacM1()
	if err != nil {
		return fmt.Errorf("failed to detect system architecture")
	}
	if isM1 {
		// TACTICAL: when installing the FDW for Mac M1, it is necessary to do a shell copy of the unzipped file
		if _, err := ungzip(sourcePath, tempdir); err != nil {
			return fmt.Errorf("could not unzip %s to %s", sourcePath, tempdir)
		}
		unzippedSoPath := filepath.Join(tempdir)
		var cpCmd = exec.Command("cp", unzippedSoPath, hubBinPath)
		if _, err := cpCmd.Output(); err != nil {
			return fmt.Errorf("could not copy extracted file %s to %s", unzippedSoPath, tempdir)
		}
	} else {
		// for other platforms, unzip directly into the destination
		if _, err := ungzip(sourcePath, hubBinPath); err != nil {
			return fmt.Errorf("could not unzip %s to %s", sourcePath, hubBinPath)
		}
	}

	fileName = image.Fdw.ControlFile
	sourcePath = filepath.Join(tempdir, fileName)
	if err := moveFileWithinPartition(sourcePath, filepath.Join(hubControlPath, fileName)); err != nil {
		return fmt.Errorf("could not install %s to %s", sourcePath, hubControlPath)
	}

	fileName = image.Fdw.SqlFile
	sourcePath = filepath.Join(tempdir, fileName)
	if err := moveFileWithinPartition(sourcePath, filepath.Join(hubSQLPath, fileName)); err != nil {
		return fmt.Errorf("could not install %s to %s", sourcePath, hubSQLPath)
	}
	return nil
}
