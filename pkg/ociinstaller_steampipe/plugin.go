package ociinstaller_steampipe

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/ociinstaller"
	"github.com/turbot/steampipe/pkg/filepaths_steampipe"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/turbot/pipe-fittings/error_helpers"
	versionfile "github.com/turbot/pipe-fittings/ociinstaller/versionfile"
	"github.com/turbot/pipe-fittings/utils"
)

var versionFileUpdateLock = &sync.Mutex{}

// InstallPlugin installs a plugin from an OCI Image
func InstallPlugin(ctx context.Context, imageRef string, sub chan struct{}, opts ...PluginInstallOption) (*ociinstaller.SteampipeImage, error) {
	config := &pluginInstallConfig{}
	for _, opt := range opts {
		opt(config)
	}
	tempDir := ociinstaller.NewTempDir(filepaths_steampipe.EnsurePluginDir())
	defer func() {
		// send a last beacon to signal completion
		sub <- struct{}{}
		if err := tempDir.Delete(); err != nil {
			log.Printf("[TRACE] Failed to delete temp dir '%s' after installing plugin: %s", tempDir, err)
		}
	}()

	ref := ociinstaller.NewSteampipeImageRef(imageRef)
	imageDownloader := ociinstaller.NewOciDownloader()

	sub <- struct{}{}
	image, err := imageDownloader.Download(ctx, ref, ociinstaller.ImageTypePlugin, tempDir.Path)
	if err != nil {
		return nil, err
	}

	sub <- struct{}{}
	if err = installPluginBinary(image, tempDir.Path); err != nil {
		return nil, fmt.Errorf("plugin installation failed: %s", err)
	}
	sub <- struct{}{}
	if err = installPluginDocs(image, tempDir.Path); err != nil {
		return nil, fmt.Errorf("plugin installation failed: %s", err)
	}
	if !config.skipConfigFile {
		if err = installPluginConfigFiles(image, tempDir.Path); err != nil {
			return nil, fmt.Errorf("plugin installation failed: %s", err)
		}
	}
	sub <- struct{}{}
	if err := updatePluginVersionFiles(image); err != nil {
		return nil, err
	}
	return image, nil
}

// updatePluginVersionFiles updates the global versions.json to add installation of the plugin
// also adds a version file in the plugin installation directory with the information
func updatePluginVersionFiles(image *ociinstaller.SteampipeImage) error {
	versionFileUpdateLock.Lock()
	defer versionFileUpdateLock.Unlock()

	timeNow := versionfile.FormatTime(time.Now())
	v, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		return err
	}

	pluginFullName := image.ImageRef.DisplayImageRef()

	installedVersion, ok := v.Plugins[pluginFullName]
	if !ok {
		installedVersion = versionfile.EmptyInstalledVersion()
	}

	installedVersion.Name = pluginFullName
	installedVersion.Version = image.Config.Plugin.Version
	installedVersion.ImageDigest = string(image.OCIDescriptor.Digest)
	installedVersion.BinaryDigest = image.Plugin.BinaryDigest
	installedVersion.BinaryArchitecture = image.Plugin.BinaryArchitecture
	installedVersion.InstalledFrom = image.ImageRef.ActualImageRef()
	installedVersion.LastCheckedDate = timeNow
	installedVersion.InstallDate = timeNow

	v.Plugins[pluginFullName] = installedVersion

	// Ensure that the version file is written to the plugin installation folder
	// Having this file is important, since this can be used
	// to compose the global version file if it is unavailable or unparseable
	// This makes sure that in the event of corruption (global/individual) we don't end up
	// losing all the plugin install data
	if err := v.EnsurePluginVersionFile(installedVersion); err != nil {
		return err
	}

	return v.Save()
}

func installPluginBinary(image *ociinstaller.SteampipeImage, tempdir string) error {
	sourcePath := filepath.Join(tempdir, image.Plugin.BinaryFile)
	destDir := pluginInstallDir(image.ImageRef)

	// check if system is M1 - if so we need some special handling
	isM1, err := utils.IsMacM1()
	if err != nil {
		return fmt.Errorf("failed to detect system architecture")
	}
	if isM1 {
		// NOTE: for Mac M1 machines, if the binary is updated in place without deleting the existing file,
		// the updated plugin binary may crash on execution - for an undetermined reason
		// to avoid this, remove the existing plugin folder and re-create it
		if err := os.RemoveAll(destDir); err != nil {
			return fmt.Errorf("could not remove plugin folder")
		}
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("could not create plugin folder")
		}
	}

	// unzip the file into the plugin folder
	if _, err := ociinstaller.Ungzip(sourcePath, destDir); err != nil {
		return fmt.Errorf("could not unzip %s to %s", sourcePath, destDir)
	}
	return nil
}

func installPluginDocs(image *ociinstaller.SteampipeImage, tempdir string) error {
	installTo := pluginInstallDir(image.ImageRef)

	// if DocsDir is not set, then there are no docs.
	if image.Plugin.DocsDir == "" {
		return nil
	}

	// install the docs
	sourcePath := filepath.Join(tempdir, image.Plugin.DocsDir)
	destPath := filepath.Join(installTo, "docs")
	if filehelpers.FileExists(destPath) {
		os.RemoveAll(destPath)
	}
	if err := ociinstaller.MoveFolderWithinPartition(sourcePath, destPath); err != nil {
		return fmt.Errorf("could not copy %s to %s", sourcePath, destPath)
	}
	return nil
}

func installPluginConfigFiles(image *ociinstaller.SteampipeImage, tempdir string) error {
	installTo := filepaths_steampipe.EnsureConfigDir()

	// if ConfigFileDir is not set, then there are no config files.
	if image.Plugin.ConfigFileDir == "" {
		return nil
	}
	// install config files (if they dont already exist)
	sourcePath := filepath.Join(tempdir, image.Plugin.ConfigFileDir)

	objects, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't read source dir: %s", err)
	}

	for _, obj := range objects {
		sourceFile := filepath.Join(sourcePath, obj.Name())
		destFile := filepath.Join(installTo, obj.Name())
		if err := copyConfigFileUnlessExists(sourceFile, destFile, image.ImageRef); err != nil {
			return fmt.Errorf("could not copy config file from %s to %s", sourceFile, destFile)
		}
	}

	return nil
}

func copyConfigFileUnlessExists(sourceFile string, destFile string, ref *ociinstaller.SteampipeImageRef) error {
	if filehelpers.FileExists(destFile) {
		return nil
	}
	inputData, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	inputStat, err := os.Stat(sourceFile)
	if err != nil {
		return fmt.Errorf("couldn't read source file permissions: %s", err)
	}
	// update the connection config with the correct plugin version
	inputData = addPluginStreamToConfig(inputData, ref)
	if err = os.WriteFile(destFile, inputData, inputStat.Mode()); err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	return nil
}

// The default config files have the plugin set to the 'latest' stream (as this is what is installed by default)
// When installing non-latest plugins, that property needs to be adjusted to the stream actually getting installed.
// Otherwise, during plugin resolution, it will resolve to an incorrect plugin instance
// (or none at at all, if  'latest' versions isn't installed)
func addPluginStreamToConfig(src []byte, ref *ociinstaller.SteampipeImageRef) []byte {
	_, _, stream := ref.GetOrgNameAndStream()
	if stream == "latest" {
		return src
	}

	regex := regexp.MustCompile(`^(\s*)plugin\s*=\s*"(.*)"\s*$`)
	substitution := fmt.Sprintf(`$1 plugin = "$2@%s"`, stream)

	srcScanner := bufio.NewScanner(strings.NewReader(string(src)))
	srcScanner.Split(bufio.ScanLines)
	destBuffer := bytes.NewBufferString("")

	for srcScanner.Scan() {
		line := srcScanner.Text()
		if regex.MatchString(line) {
			line = regex.ReplaceAllString(line, substitution)
			// remove the extra space we had to add to the substitution token
			line = line[1:]
		}
		destBuffer.WriteString(fmt.Sprintf("%s\n", line))
	}
	return destBuffer.Bytes()
}

func pluginInstallDir(ref *ociinstaller.SteampipeImageRef) string {
	osSafePath := filepath.FromSlash(ref.DisplayImageRef())

	fullPath := filepath.Join(filepaths_steampipe.EnsurePluginDir(), osSafePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		err = os.MkdirAll(fullPath, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create plugin install directory")
	}

	return fullPath
}
