package ociinstaller

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
	"github.com/turbot/steampipe/pkg/utils"
)

var versionFileUpdateLock = &sync.Mutex{}

// InstallPlugin installs a plugin from an OCI Image
func InstallPlugin(ctx context.Context, imageRef string, constraint string, sub chan struct{}, opts ...PluginInstallOption) (*SteampipeImage, error) {
	config := &pluginInstallConfig{}
	for _, opt := range opts {
		opt(config)
	}
	tempDir := NewTempDir(filepaths.EnsurePluginDir())
	defer func() {
		// send a last beacon to signal completion
		sub <- struct{}{}
		if err := tempDir.Delete(); err != nil {
			log.Printf("[TRACE] Failed to delete temp dir '%s' after installing plugin: %s", tempDir, err)
		}
	}()

	ref := NewSteampipeImageRef(imageRef)
	imageDownloader := NewOciDownloader()

	sub <- struct{}{}
	image, err := imageDownloader.Download(ctx, ref, ImageTypePlugin, tempDir.Path)
	if err != nil {
		return nil, err
	}

	pluginPath := pluginInstallDir(image.ImageRef, constraint)

	sub <- struct{}{}
	if err = installPluginBinary(image, tempDir.Path, pluginPath); err != nil {
		return nil, fmt.Errorf("plugin installation failed: %s", err)
	}
	sub <- struct{}{}
	if err = installPluginDocs(image, tempDir.Path, pluginPath); err != nil {
		return nil, fmt.Errorf("plugin installation failed: %s", err)
	}
	if !config.skipConfigFile {
		if err = installPluginConfigFiles(image, tempDir.Path, constraint); err != nil {
			return nil, fmt.Errorf("plugin installation failed: %s", err)
		}
	}
	sub <- struct{}{}
	if err := updatePluginVersionFiles(ctx, image, constraint); err != nil {
		return nil, err
	}
	return image, nil
}

// updatePluginVersionFiles updates the global versions.json to add installation of the plugin
// also adds a version file in the plugin installation directory with the information
func updatePluginVersionFiles(ctx context.Context, image *SteampipeImage, constraint string) error {
	versionFileUpdateLock.Lock()
	defer versionFileUpdateLock.Unlock()

	timeNow := versionfile.FormatTime(time.Now())
	v, err := versionfile.LoadPluginVersionFile(ctx)
	if err != nil {
		return err
	}

	// For the full name we want the constraint (^0.4) used, not the resolved version (0.4.1)
	// we override the DisplayImageRef with the constraint here.
	pluginFullName := image.ImageRef.DisplayImageRefConstraintOverride(constraint)

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

func installPluginBinary(image *SteampipeImage, tempdir string) error {
	sourcePath := filepath.Join(tempdir, image.Plugin.BinaryFile)
	destDir := filepaths.EnsurePluginInstallDir(image.ImageRef.DisplayImageRef())

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
	if _, err := ungzip(sourcePath, destDir); err != nil {
		return fmt.Errorf("could not unzip %s to %s", sourcePath, pluginDir)
	}
	return nil
}

func installPluginDocs(image *SteampipeImage, tempdir string) error {
	installTo := filepaths.EnsurePluginInstallDir(image.ImageRef.DisplayImageRef())

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

func installPluginConfigFiles(image *SteampipeImage, tempdir string, constraint string) error {
	installTo := filepaths.EnsureConfigDir()

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
		if err := copyConfigFileUnlessExists(sourceFile, destFile, constraint); err != nil {
			return fmt.Errorf("could not copy config file from %s to %s", sourceFile, destFile)
		}
	}

	return nil
}

func copyConfigFileUnlessExists(sourceFile string, destFile string, constraint string) error {
	if fileExists(destFile) {
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
	inputData = addPluginConstraintToConfig(inputData, constraint)
	if err = os.WriteFile(destFile, inputData, inputStat.Mode()); err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	return nil
}

// The default config files have the plugin set to the 'latest' stream (as this is what is installed by default)
// When installing non-latest plugins, that property needs to be adjusted to the stream actually getting installed.
// Otherwise, during plugin resolution, it will resolve to an incorrect plugin instance
// (or none at all, if  'latest' versions isn't installed)
func addPluginConstraintToConfig(src []byte, constraint string) []byte {
	if constraint == "latest" {
		return src
	}

	regex := regexp.MustCompile(`^(\s*)plugin\s*=\s*"(.*)"\s*$`)
	substitution := fmt.Sprintf(`$1 plugin = "$2@%s"`, constraint)

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

func pluginInstallDir(ref *SteampipeImageRef, constraint string) string {
	osSafePath := filepath.FromSlash(ref.DisplayImageRefConstraintOverride(constraint))
	fullPath := filepath.Join(filepaths.EnsurePluginDir(), osSafePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		err = os.MkdirAll(fullPath, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create plugin install directory")
	}

	return fullPath
}
