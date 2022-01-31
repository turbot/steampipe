package ociinstaller

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
)

// InstallAssets installs the Steampipe report server assets
func InstallAssets(ctx context.Context, dest string) error {
	imageRef := constants.AssetsImageRef
	tempDir := NewTempDir(imageRef)
	defer tempDir.Delete()

	// download the blobs.
	imageDownloader := NewOciDownloader()
	image, err := imageDownloader.Download(ctx, imageRef, ImageTypeAssets, tempDir.Path)
	if err != nil {
		return err
	}

	// install the files
	if err = installAssetsFiles(image, tempDir.Path, dest); err != nil {
		return err
	}

	return nil
}

func installAssetsFiles(image *SteampipeImage, tempdir string, dest string) error {
	fileName := image.Assets.ReportUI
	sourcePath := filepath.Join(tempdir, fileName)
	if err := moveFolderWithinPartition(sourcePath, filepaths.ReportAssetsPath()); err != nil {
		return fmt.Errorf("could not install %s to %s", sourcePath, filepaths.ReportAssetsPath())
	}
	return nil
}
