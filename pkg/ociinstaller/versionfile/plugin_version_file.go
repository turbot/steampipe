package versionfile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
)

var (
	ErrNoContent = errors.New("no content")
)

const (
	PluginStructVersion = 20220411
	// the name of the version files that are put in the plugin installation directories
	pluginVersionFileName = "version.json"
)

type PluginVersionFile struct {
	Plugins       map[string]*InstalledVersion `json:"plugins"`
	StructVersion int64                        `json:"struct_version"`
}

func newPluginVersionFile() *PluginVersionFile {
	return &PluginVersionFile{
		Plugins:       map[string]*InstalledVersion{},
		StructVersion: PluginStructVersion,
	}
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (p *PluginVersionFile) IsValid() bool {
	return p.StructVersion > 0
}

// EnsurePluginVersionFile reads the version file in the plugin directory (if exists) and overwrites it if the data in the
// argument is different. The comparison is done using the `Name` and `BinaryDigest` properties.
// If the file doesn't exist, or cannot be read/parsed, EnsurePluginVersionFile fails over to overwriting the data
func (p *PluginVersionFile) EnsurePluginVersionFile(installData *InstalledVersion) error {
	pluginFolder, err := filepaths.FindPluginFolder(installData.Name)
	if err != nil {
		return err
	}
	versionFile := filepath.Join(pluginFolder, pluginVersionFileName)

	// If the version file already exists, we only write to it if the incoming data is newer
	if filehelpers.FileExists(versionFile) {
		installation, err := readPluginVersionFile(versionFile)
		if err == nil && installation.Equal(installData) {
			// the new and old data match - no need to overwrite
			return nil
		}
		// in case of error, just failover to a overwrite
	}

	theBytes, err := json.MarshalIndent(installData, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(versionFile, theBytes, 0644)
}

// Save writes the config file to disk
func (p *PluginVersionFile) Save() error {
	// set struct version
	p.StructVersion = PluginStructVersion
	versionFilePath := filepaths.PluginVersionFilePath()
	return p.write(versionFilePath)
}

func (p *PluginVersionFile) write(path string) error {
	versionFileJSON, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		log.Println("[ERROR]", "Error while writing version file", err)
		return err
	}
	if len(versionFileJSON) == 0 {
		log.Println("[ERROR]", "Cannot write 0 bytes to file")
		return sperr.WrapWithMessage(ErrNoContent, "cannot write versions file")
	}
	return os.WriteFile(path, versionFileJSON, 0644)
}

func (p *PluginVersionFile) ensureVersionFilesInPluginDirectories() error {
	removals := []*InstalledVersion{}
	for _, installation := range p.Plugins {
		if err := p.EnsurePluginVersionFile(installation); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				removals = append(removals, installation)
				continue
			}
			return err
		}
	}

	// if we found any plugins that do not have installations, remove them from the map
	if len(removals) > 0 {
		for _, removal := range removals {
			delete(p.Plugins, removal.Name)
		}
		return p.Save()
	}
	return nil
}

// any plugins installed under the `local` folder are added to the plugin version file
func (p *PluginVersionFile) AddLocalPlugins(ctx context.Context) *error_helpers.ErrorAndWarnings {
	localPlugins, err := loadLocalPlugins(ctx)
	if err != nil {
		return error_helpers.NewErrorsAndWarning(err)
	}
	for name, install := range localPlugins {
		if _, ok := p.Plugins[name]; ok {
			// if the plugin is already in the global version file, skip it
			continue
		}
		p.Plugins[fmt.Sprintf("local/%s", name)] = install
	}
	return nil
}

// to lock plugin version file loads
var pluginLoadLock = sync.Mutex{}

// LoadPluginVersionFile migrates from the old version file format if necessary and loads the plugin version data
func LoadPluginVersionFile(ctx context.Context) (*PluginVersionFile, error) {

	// we need a lock here so that we don't hit a race condition where
	// the plugin file needs to be composed
	// if recomposition is not required, this has (almost) zero penalty
	pluginLoadLock.Lock()
	defer pluginLoadLock.Unlock()

	versionFilePath := filepaths.PluginVersionFilePath()
	if filehelpers.FileExists(versionFilePath) {
		pluginVersions, err := readGlobalPluginVersionsFile(versionFilePath)

		// we could read and parse out the file - all is well
		if err == nil {
			return pluginVersions, nil
		}
	}

	// we don't have a global plugin/versions.json or it is not parseable or is empty (always recompose)
	// generate the version file from the individual version files by walking the plugin directories
	// this will return an Empty Version file if there are no version files in the plugin directories
	pluginVersions := recomposePluginVersionFile(ctx)

	// save the recomposed file
	err := pluginVersions.Save()
	if err != nil {
		return nil, err
	}
	return pluginVersions, err
}

func loadLocalPlugins(ctx context.Context) (map[string]*InstalledVersion, error) {
	localFolder := filepaths.LocalPluginPath()
	localPlugins := map[string]*InstalledVersion{}

	// iterate over all folders underneath the local plugin directory and if the folder contains a plugin, add to the map
	pluginFolders, err := filehelpers.ListFilesWithContext(ctx, localFolder, &filehelpers.ListOptions{Flags: filehelpers.DirectoriesFlat})
	if err != nil {
		return nil, err
	}
	for _, pluginFolder := range pluginFolders {
		// check if the folder contains a plugin file
		pluginName := filepath.Base(pluginFolder)
		pluginFile := filepath.Join(pluginFolder, pluginName+".plugin")
		if filehelpers.FileExists(pluginFile) {
			// read the plugin file

			localPlugins[pluginName] = &InstalledVersion{
				Name:          pluginName,
				Version:       "latest",
				StructVersion: InstalledVersionStructVersion,
			}
		}
	}

	return localPlugins, nil
}

// EnsureVersionFilesInPluginDirectories attempts a backfill of the individual version.json for plugins
// this is required only once when upgrading from 0.20.x
func EnsureVersionFilesInPluginDirectories(ctx context.Context) error {
	versions, err := LoadPluginVersionFile(ctx)
	if err != nil {
		return err
	}
	return versions.ensureVersionFilesInPluginDirectories()
}

// recomposePluginVersionFile recursively traverses down the plugin direcory and tries to
// recompose the global version file from the plugin version files
// if there are no plugin version files, this returns a ready to use empty global version file
func recomposePluginVersionFile(ctx context.Context) *PluginVersionFile {
	pvf := newPluginVersionFile()

	versionFiles, err := filehelpers.ListFilesWithContext(ctx, filepaths.EnsurePluginDir(), &filehelpers.ListOptions{
		Include: []string{fmt.Sprintf("**/%s", pluginVersionFileName)},
		Flags:   filehelpers.FilesRecursive,
	})

	if err != nil {
		log.Println("[TRACE] recomposePluginVersionFile failed - error while walking plugin directory for version files", err)
		return pvf
	}

	for _, versionFile := range versionFiles {
		install, err := readPluginVersionFile(versionFile)
		if err != nil {
			log.Println("[TRACE] could not read file", versionFile)
			continue
		}
		pvf.Plugins[install.Name] = install
	}

	return pvf
}

func readPluginVersionFile(versionFile string) (*InstalledVersion, error) {
	data, err := os.ReadFile(versionFile)
	if err != nil {
		log.Println("[TRACE] could not read file", versionFile)
		return nil, err
	}
	install := EmptyInstalledVersion()
	if err := json.Unmarshal(data, &install); err != nil {
		// this wasn't the version file (probably) - keep going
		log.Println("[TRACE] unmarshal failed for file:", versionFile)
		return nil, err
	}
	return install, nil
}

func readGlobalPluginVersionsFile(path string) (*PluginVersionFile, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(file) == 0 {
		// the file exists, but is empty - return an error
		// start from scratch
		return nil, sperr.New("plugin versions.json file is empty")
	}

	var data PluginVersionFile

	if err := json.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	if data.Plugins == nil {
		data.Plugins = map[string]*InstalledVersion{}
	}

	for key, installedPlugin := range data.Plugins {
		// hard code the name to the key
		installedPlugin.Name = key
		if installedPlugin.StructVersion == 0 {
			// also backfill the StructVersion in map values
			installedPlugin.StructVersion = InstalledVersionStructVersion
		}
	}

	return &data, nil
}
