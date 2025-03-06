package ociinstaller

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/steampipe/pkg/constants"
)

// InstallAssets installs the Steampipe report server assets
func InstallAssets(ctx context.Context, assetsLocation string) error {
	tempDir := ociinstaller.NewTempDir(assetsLocation)
	defer func() {
		if err := tempDir.Delete(); err != nil {
			log.Printf("[TRACE] Failed to delete temp dir '%s' after installing assets: %s", tempDir, err)
		}
	}()

	// download the blobs
	imageDownloader := newAssetDownloader()
	image, err := imageDownloader.Download(ctx, ociinstaller.NewImageRef(constants.DashboardAssetsImageRef), ImageTypeAssets, tempDir.Path)
	if err != nil {
		return err
	}

	// install the files
	if err = installAssetsFiles(image, tempDir.Path, assetsLocation); err != nil {
		return err
	}

	return nil
}

func installAssetsFiles(image *ociinstaller.OciImage[*assetsImage, *assetsImageConfig], tempdir string, dest string) error {
	fileName := image.Data.ReportUI
	sourcePath := filepath.Join(tempdir, fileName)
	if err := ociinstaller.MoveFolderWithinPartition(sourcePath, dest); err != nil {
		return fmt.Errorf("could not install %s to %s", sourcePath, dest)
	}
	return nil
}
