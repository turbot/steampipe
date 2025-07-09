package ociinstaller

import (
	"context"
	"fmt"
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
		return string(image.OCIDescriptor.Digest), err
	}

	return string(image.OCIDescriptor.Digest), nil
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
	fdwBinDir := filepaths.GetFDWBinaryDir()
	fdwBinFileSourcePath := filepath.Join(tempdir, image.Data.BinaryFile)
	fdwBinFileDestPath := filepath.Join(fdwBinDir, constants.FdwBinaryFileName)

	// NOTE: for Mac M1 machines, if the fdw binary is updated in place without deleting the existing file,
	// the updated fdw may crash on execution - for an undetermined reason
	// to avoid this, first remove the existing .so file
	os.Remove(fdwBinFileDestPath)
	// now unzip the fdw file
	if _, err := ociinstaller.Ungzip(fdwBinFileSourcePath, fdwBinDir); err != nil {
		return fmt.Errorf("could not unzip %s to %s: %s", fdwBinFileSourcePath, fdwBinDir, err.Error())
	}

	fdwControlDir := filepaths.GetFDWSQLAndControlDir()
	controlFileName := image.Data.ControlFile
	controlFileSourcePath := filepath.Join(tempdir, controlFileName)
	controlFileDestPath := filepath.Join(fdwControlDir, image.Data.ControlFile)

	if err := ociinstaller.MoveFileWithinPartition(controlFileSourcePath, controlFileDestPath); err != nil {
		return fmt.Errorf("could not install %s to %s", controlFileSourcePath, fdwControlDir)
	}

	fdwSQLDir := filepaths.GetFDWSQLAndControlDir()
	sqlFileName := image.Data.SqlFile
	sqlFileSourcePath := filepath.Join(tempdir, sqlFileName)
	sqlFileDestPath := filepath.Join(fdwSQLDir, sqlFileName)
	if err := ociinstaller.MoveFileWithinPartition(sqlFileSourcePath, sqlFileDestPath); err != nil {
		return fmt.Errorf("could not install %s to %s", sqlFileSourcePath, fdwSQLDir)
	}
	return nil
}
