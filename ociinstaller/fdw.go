package ociinstaller

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	versionfile "github.com/turbot/steampipe/ociinstaller/versionfile"
)

// InstallFdw :: Install the Steampipe Hub Extension files from an OCI image
func InstallFdw(ctx context.Context, imageRef string, dbLocation string) (string, error) {
	tempDir := NewTempDir(imageRef)
	defer tempDir.Delete()

	imageDownloader := NewOciDownloader()

	// download the blobs.
	image, err := imageDownloader.Download(ctx, imageRef, "fdw", tempDir.Path)
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
	v.FdwExtension.InstalledFrom = image.ImageRef
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
	if _, err := ungzip(sourcePath, hubBinPath); err != nil {
		return fmt.Errorf("could not unzip %s to %s", sourcePath, hubBinPath)
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
