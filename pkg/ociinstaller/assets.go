package ociinstaller

import (
	"context"
	"fmt"
	"path/filepath"
	"log"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
)

// InstallAssets installs the Steampipe report server assets
func InstallAssets(ctx context.Context, assetsLocation string) error {
	tempDir := NewTempDir(assetsLocation)
	defer func() {
		if err := tempDir.Delete(); err != nil {
			log.Printf("[TRACE] Failed to delete temp dir '%s' after installing assets: %s", tempDir, err)
		}
	}()

	// download the blobs
	imageDownloader := NewOciDownloader()
	image, err := imageDownloader.Download(ctx, NewSteampipeImageRef(constants.DashboardAssetsImageRef), ImageTypeAssets, tempDir.Path)
	if err != nil {
		return err
	}

	// install the files
	if err = installAssetsFiles(image, tempDir.Path, assetsLocation); err != nil {
		return err
	}

	return nil
}

func installAssetsFiles(image *SteampipeImage, tempdir string, dest string) error {
	fileName := image.Assets.ReportUI
	sourcePath := filepath.Join(tempdir, fileName)
	if err := moveFolderWithinPartition(sourcePath, filepaths.EnsureDashboardAssetsDir()); err != nil {
		return fmt.Errorf("could not install %s to %s", sourcePath, filepaths.EnsureDashboardAssetsDir())
	}
	return nil
}
