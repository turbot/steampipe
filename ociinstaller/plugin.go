package ociinstaller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/constants"
	versionfile "github.com/turbot/steampipe/ociinstaller/versionfile"
	"github.com/turbot/steampipe/utils"
)

// InstallPlugin installs a plugin from an OCI Image
func InstallPlugin(ctx context.Context, imageRef string) (*SteampipeImage, error) {
	tempDir := NewTempDir(imageRef)
	defer tempDir.Delete()

	ref := NewSteampipeImageRef(imageRef)
	imageDownloader := NewOciDownloader()

	image, err := imageDownloader.Download(ctx, ref.ActualImageRef(), "plugin", tempDir.Path)
	if err != nil {
		return nil, err
	}

	if err = installPluginBinary(image, tempDir.Path); err != nil {
		return nil, fmt.Errorf("plugin installation failed: %s", err)
	}
	if err = installPluginDocs(image, tempDir.Path); err != nil {
		return nil, fmt.Errorf("plugin installation failed: %s", err)
	}
	if err = installPluginConfigFiles(image, tempDir.Path); err != nil {
		return nil, fmt.Errorf("plugin installation failed: %s", err)
	}

	if err := updateVersionFilePlugin(image); err != nil {
		return nil, err
	}
	return image, nil
}

func updateVersionFilePlugin(image *SteampipeImage) error {
	timeNow := versionfile.FormatTime(time.Now())
	v, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		return err
	}

	ref := NewSteampipeImageRef(image.ImageRef)

	pluginFullName := ref.DisplayImageRef()

	plugin, ok := v.Plugins[pluginFullName]
	if ok == false {
		plugin = &versionfile.InstalledVersion{}
	}

	//change this to the path????
	plugin.Name = pluginFullName
	plugin.Version = image.Config.Plugin.Version
	plugin.ImageDigest = string(image.OCIDescriptor.Digest)
	plugin.InstalledFrom = ref.ActualImageRef()
	plugin.LastCheckedDate = timeNow
	plugin.InstallDate = timeNow

	v.Plugins[pluginFullName] = plugin

	return v.Save()
}

func installPluginBinary(image *SteampipeImage, tempdir string) error {
	installTo := pluginInstallDir(image.ImageRef)

	// install the binary file
	fileName := image.Plugin.BinaryFile
	sourcePath := filepath.Join(tempdir, fileName)
	if _, err := ungzip(sourcePath, installTo); err != nil {
		return fmt.Errorf("could not unzip %s to %s", sourcePath, installTo)
	}

	return nil
}

func installPluginDocs(image *SteampipeImage, tempdir string) error {
	installTo := pluginInstallDir(image.ImageRef)

	// if DocsDir is not set, then there are no docs.
	if image.Plugin.DocsDir == "" {
		return nil
	}

	// install the docs
	sourcePath := filepath.Join(tempdir, image.Plugin.DocsDir)
	destPath := filepath.Join(installTo, "docs")
	if fileExists(destPath) {
		os.RemoveAll(destPath)
	}
	if err := moveFolderWithinPartition(sourcePath, destPath); err != nil {
		return fmt.Errorf("could not copy %s to %s", sourcePath, destPath)
	}
	return nil
}

func installPluginConfigFiles(image *SteampipeImage, tempdir string) error {
	installTo := constants.ConfigDir()

	// if ConfigFileDir is not set, then there are no config files.
	if image.Plugin.ConfigFileDir == "" {
		return nil
	}
	// install config files (if they dont already exist)
	sourcePath := filepath.Join(tempdir, image.Plugin.ConfigFileDir)
	directory, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source dir: %s", err)
	}
	defer directory.Close()

	objects, err := directory.Readdir(-1)
	if err != nil {
		return fmt.Errorf("couldn't read source dir: %s", err)
	}

	for _, obj := range objects {
		sourceFile := filepath.Join(sourcePath, obj.Name())
		destFile := filepath.Join(installTo, obj.Name())

		if err := copyFileUnlessExists(sourceFile, destFile); err != nil {
			return fmt.Errorf("could not copy %s to %s", sourceFile, destFile)

		}
	}

	return nil
}

func pluginInstallDir(imageRef string) string {
	ref := NewSteampipeImageRef(imageRef)
	osSafePath := filepath.FromSlash(ref.DisplayImageRef())

	fullPath := filepath.Join(constants.PluginDir(), osSafePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		err = os.MkdirAll(fullPath, 0755)
		utils.FailOnErrorWithMessage(err, "could not create plugin install directory")
	}

	return fullPath
}
