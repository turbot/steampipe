package versionfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/migrate"
	"github.com/turbot/steampipe/sperr"
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

	// used when this structure is populated by traversing individual plugin directory
	// if so, there's no need to attempt a backfill
	composed bool
}

func newPluginVersionFile() *PluginVersionFile {
	return &PluginVersionFile{
		Plugins:       map[string]*InstalledVersion{},
		StructVersion: PluginStructVersion,
	}
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (f PluginVersionFile) IsValid() bool {
	return f.StructVersion > 0
}

func (f *PluginVersionFile) MigrateFrom() migrate.Migrateable {
	f.StructVersion = PluginStructVersion
	for p := range f.Plugins {
		f.Plugins[p].MigrateLegacy()
	}
	return f
}

func (f *PluginVersionFile) EnsureVersionFile(installData *InstalledVersion, force bool) error {
	pluginFolder, err := filepaths.FindPluginFolder(installData.Name)
	if err != nil {
		return err
	}
	versionFile := filepath.Join(pluginFolder, pluginVersionFileName)

	// if this is not forced, make sure that the file doesn't exist before overwriting it
	if !force && filehelpers.FileExists(versionFile) {
		return nil
	}

	// make sure that the legacy fields are also filled in
	installData.MaintainLegacy()

	theBytes, err := json.MarshalIndent(installData, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(versionFile, theBytes, 0644)
}

// Save writes the config file to disk
func (f *PluginVersionFile) Save() error {
	// set struct version
	f.StructVersion = PluginStructVersion
	versionFilePath := filepaths.PluginVersionFilePath()
	// maintain the legacy properties for backward compatibility
	for _, v := range f.Plugins {
		v.MaintainLegacy()
	}
	return f.write(versionFilePath)
}

func (f *PluginVersionFile) write(path string) error {
	versionFileJSON, err := json.MarshalIndent(f, "", "  ")
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

func (f *PluginVersionFile) backfill() error {
	for _, installation := range f.Plugins {
		if err := f.EnsureVersionFile(installation, false); err != nil {
			return err
		}
	}
	return nil
}

// to lock plugin version file loads
var pluginLoadLock = sync.Mutex{}

// LoadPluginVersionFile migrates from the old version file format if necessary and loads the plugin version data
func LoadPluginVersionFile() (*PluginVersionFile, error) {
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

		// check if this was a syntax error during parsing
		// if it is a syntax error, either the file is corrupted or empty
		var syntaxError *json.SyntaxError
		isSyntaxError := errors.As(err, &syntaxError)
		if !isSyntaxError {
			// not a syntax error - return as-is
			return nil, err
		}
	}

	// we don't have a global plugin/versions.json or it is not parseable
	// generate the version file from the individual version files by walking the plugin directories
	// this will return an Empty Version file if there are no version files in the plugin directories
	pluginVersions := recomposePluginVersionFile()

	// save the recomposed file
	err := pluginVersions.Save()
	if err != nil {
		return nil, err
	}
	return pluginVersions, err
}

// BackfillPluginVersionFile attempts a backfill of the individual version.json for plugins
// this is required only once when upgrading from 0.20.x
func BackfillPluginVersionFile() error {
	versions, err := LoadPluginVersionFile()
	if err != nil {
		return err
	}
	if versions.composed {
		// this was composed from the plugin directories
		// no point backfilling
		return nil
	}
	return versions.backfill()
}

// recomposePluginVersionFile recursively traverses down the plugin direcory and tries to
// recompose the global version file from the plugin version files
// if there are no plugin version files, this returns a ready to use empty global version file
func recomposePluginVersionFile() *PluginVersionFile {
	pvf := newPluginVersionFile()

	versionFiles, err := filehelpers.ListFiles(filepaths.EnsurePluginDir(), &filehelpers.ListOptions{
		Include: []string{fmt.Sprintf("**/%s", pluginVersionFileName)},
		Flags:   filehelpers.FilesRecursive,
	})

	if err != nil {
		log.Println("[TRACE] recomposePluginVersionFile failed - error while walking plugin directory for version files", err)
		return pvf
	}

	for _, versionFile := range versionFiles {
		data, err := os.ReadFile(versionFile)
		if err != nil {
			log.Println("[TRACE] could not read file", versionFile)
			continue
		}
		install := EmptyInstalledVersion()
		if err := json.Unmarshal(data, &install); err != nil {
			// this wasn't the version file (probably) - keep going
			log.Println("[TRACE] unmarshal failed for file:", versionFile)
			continue
		}
		pvf.Plugins[install.Name] = install

		// mark that this is a composed version file
		// and not just read
		pvf.composed = true
	}

	return pvf
}

func readGlobalPluginVersionsFile(path string) (*PluginVersionFile, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(file) == 0 {
		// the file exists, but is empty
		// start from scratch
		return newPluginVersionFile(), nil
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
