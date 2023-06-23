package versionfile

import (
	"encoding/json"
	"errors"
	"io/fs"
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
	PluginStructVersion   = 20220411
	pluginVersionFileName = "version.json"
)

type PluginVersionFile struct {
	Plugins       map[string]*InstalledVersion `json:"plugins"`
	StructVersion int64                        `json:"struct_version"`

	// used when this structure was populated by traversion plugin directory
	// if so, there's no need to attempt a backfill
	composed bool
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

func (f *PluginVersionFile) Backfill() error {
	for _, installation := range f.Plugins {
		if err := f.EnsureVersionFile(installation, false); err != nil {
			return err
		}
	}
	return nil
}

func (f *PluginVersionFile) EnsureVersionFile(installData *InstalledVersion, force bool) error {
	pluginFolder, err := filepaths.FindPluginFolder(installData.Name)
	if err != nil {
		return err
	}
	versionFile := filepath.Join(pluginFolder, pluginVersionFileName)
	if !force {
		// if this is not forced, make sure that the file doesn't exist before overwriting it
		if filehelpers.FileExists(versionFile) {
			return nil
		}
	}

	// make sure that the legacy fields are also filled in
	installData.MaintainLegacy()

	theBytes, err := json.MarshalIndent(installData, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(versionFile, theBytes, 0644)
}

func NewPluginVersionFile() *PluginVersionFile {
	return &PluginVersionFile{
		Plugins:       map[string]*InstalledVersion{},
		StructVersion: PluginStructVersion,
	}
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
		pluginVersions, err := readPluginVersionFile(versionFilePath)

		// we could read out the file
		if err == nil {
			return pluginVersions, nil
		}

		// check if this was a syntax error during parsing
		var syntaxError *json.SyntaxError
		isSyntaxError := errors.As(err, &syntaxError)
		if !isSyntaxError {
			// no - return
			return nil, err
		}

		// generate the version file from the individual version files
		// by walking the plugin directories
		pluginVersions = recomposePluginVersionFile()
		err = pluginVersions.Save()

		return pluginVersions, err
	}

	return NewPluginVersionFile(), nil
}

func recomposePluginVersionFile() *PluginVersionFile {
	pvf := NewPluginVersionFile()
	pluginDir := filepaths.EnsurePluginDir()

	err := filepath.WalkDir(pluginDir, func(path string, d fs.DirEntry, _ error) error {
		if !d.IsDir() {
			return nil
		}
		versionFile := filepath.Join(path, pluginVersionFileName)
		if !filehelpers.FileExists(versionFile) {
			return nil
		}
		data, err := os.ReadFile(versionFile)
		if err != nil {
			return nil
		}
		install := EmptyInstalledVersion()
		if err := json.Unmarshal(data, &install); err != nil {
			// this wasn't the version file (probably) - keep going
			return nil
		}
		pvf.Plugins[install.Name] = install
		// now that we have an entry - lets skip sub directories
		return fs.SkipDir
	})

	if err != nil {
		log.Println("[ERROR]", "error while walking plugin directory for version files", err)
	}

	// mark that this is a composed version file
	// and not directly read
	pvf.composed = true
	return pvf
}

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
	return versions.Backfill()
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

func readPluginVersionFile(path string) (*PluginVersionFile, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(file) == 0 {
		// the file exists, but is empty
		// start from scratch
		return NewPluginVersionFile(), nil
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
			// also backfill the StructVersion in the values
			installedPlugin.StructVersion = InstalledVersionStructVersion
		}
	}

	return &data, nil
}
