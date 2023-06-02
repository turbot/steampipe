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

const PluginStructVersion = 20220411

type PluginVersionFile struct {
	Plugins       map[string]*InstalledVersion `json:"plugins"`
	StructVersion int64                        `json:"struct_version"`
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
		if err := f.EnsureVersionFile(installation); err != nil {
			return err
		}
	}
	return nil
}

func (f *PluginVersionFile) EnsureVersionFile(installData *InstalledVersion) error {
	pluginFolder, err := filepaths.FindPluginFolder(installData.Name)
	if err != nil {
		return err
	}
	versionFile := filepath.Join(pluginFolder, "version.json")
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
		if err == nil {
			// backfill the InstalledVersion struct version
			// so that when this gets saved, the struct versions are filled in
			for _, iv := range pluginVersions.Plugins {
				if iv.StructVersion == 0 {
					iv.StructVersion = InstalledVersionStructVersion
				}
			}
			if pluginVersions.Plugins == nil {
				// generate the version file from the individual version files
				pluginVersions = recomposePluginVersionFile()
				pluginVersions.Save()
			}
			return pluginVersions, nil
		}
		if errors.Is(err, &json.SyntaxError{}) {
			// generate the version file from the individual version files
			pluginVersions = recomposePluginVersionFile()
			pluginVersions.Save()
		}
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
		versionFile := filepath.Join(path, "version.json")
		if !filehelpers.FileExists(versionFile) {
			return nil
		}
		data, err := os.ReadFile(versionFile)
		if err != nil {
			log.Println("[ERROR]", "could not read plugin version file at", versionFile, err)
			return nil
		}
		install := EmptyInstalledVersion()
		if err := json.Unmarshal(data, &install); err != nil {
			log.Println("[ERROR]", "error while parsing plugin version file at", versionFile, err)
			return nil
		}
		pvf.Plugins[install.Name] = install
		return nil
	})

	if err != nil {
		log.Println("[ERROR]", "error while walking plugin directory for version files", err)
	}

	return pvf
}

func BackfillPluginVersionFile() error {
	versions, err := LoadPluginVersionFile()
	if err != nil {
		return err
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
		log.Println("[ERROR]", "Error while reading plugin version file", err)
		return nil, err
	}

	if data.Plugins == nil {
		data.Plugins = map[string]*InstalledVersion{}
	}

	for key := range data.Plugins {
		// hard code the name to the key
		data.Plugins[key].Name = key
	}

	return &data, nil
}
