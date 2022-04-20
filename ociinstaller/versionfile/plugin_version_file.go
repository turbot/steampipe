package versionfile

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/migrate"
	"github.com/turbot/steampipe/utils"
)

const PluginStructVersion = 20220411

var versionFileMutex = sync.Mutex{}

// LegacyPluginVersionFile is a struct used to migrate the
// PluginVersionFile to serialize with snake case property names(migrated in v0.14.0)
type LegacyPluginVersionFile struct {
	Plugins map[string]*LegacyInstalledVersion `json:"plugins"`
}

type PluginVersionFile struct {
	mtx           *sync.Mutex
	Plugins       map[string]*InstalledVersion `json:"plugins"`
	StructVersion int64                        `json:"struct_version"`
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (f PluginVersionFile) IsValid() bool {
	return f.StructVersion > 0
}

func (f *PluginVersionFile) MigrateFrom(prev interface{}) migrate.Migrateable {
	legacyState := prev.(LegacyPluginVersionFile)
	f.StructVersion = PluginStructVersion
	f.Plugins = make(map[string]*InstalledVersion, len(legacyState.Plugins))
	f.mtx = &versionFileMutex

	for p, i := range legacyState.Plugins {
		f.Plugins[p] = &InstalledVersion{
			Name:            i.Name,
			Version:         i.Version,
			ImageDigest:     i.ImageDigest,
			InstalledFrom:   i.InstalledFrom,
			LastCheckedDate: i.LastCheckedDate,
			InstallDate:     i.InstallDate,
		}
	}
	return f
}

func NewPluginVersionFile() *PluginVersionFile {
	return &PluginVersionFile{
		mtx:           &versionFileMutex,
		Plugins:       map[string]*InstalledVersion{},
		StructVersion: PluginStructVersion,
	}
}

func pluginVersionFileFromLegacy(legacyFile *LegacyCompositeVersionFile) *PluginVersionFile {
	return &PluginVersionFile{
		Plugins: legacyFile.Plugins,
	}
}

// LoadPluginVersionFile migrates from the old version file format if necessary and loads the plugin version data
func LoadPluginVersionFile() (*PluginVersionFile, error) {
	versionFilePath := filepaths.PluginVersionFilePath()
	if helpers.FileExists(versionFilePath) {
		return readPluginVersionFile(versionFilePath)
	}
	return NewPluginVersionFile(), nil
}

// Save writes the config file to disk
func (f *PluginVersionFile) Save() ([]byte, error) {
	// set struct version
	f.StructVersion = PluginStructVersion
	versionFilePath := filepaths.PluginVersionFilePath()
	return f.write(versionFilePath)
}

func (f *PluginVersionFile) write(path string) ([]byte, error) {
	if f.mtx == nil {
		f.mtx = &versionFileMutex
	}
	f.mtx.Lock()
	defer f.mtx.Unlock()

	versionFileJSON, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		log.Println("[ERROR]", "Error while writing version file", err)
		return nil, err
	}
	return versionFileJSON, os.WriteFile(path, versionFileJSON, 0644)
}

func (f *PluginVersionFile) validate() []error {
	errFound := []error{}
	for _, iv := range f.Plugins {
		if len(iv.ImageDigest) == 0 {
			errFound = append(errFound, fmt.Errorf(iv.Name))
		}
	}
	return errFound
}

// delete the file on disk if it exists
func (f *PluginVersionFile) delete() {
	versionFilePath := filepaths.PluginVersionFilePath()
	if helpers.FileExists(versionFilePath) {
		os.Remove(versionFilePath)
	}
}

func readPluginVersionFile(path string) (*PluginVersionFile, error) {
	file, _ := os.ReadFile(path)

	var data PluginVersionFile

	if err := json.Unmarshal([]byte(file), &data); err != nil {
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

	if validationErrors := data.validate(); len(validationErrors) > 0 {
		// the data in the file is corrupt
		for _, e := range validationErrors {
			log.Printf("[WARN] information for %s is invalid in %s\n", e.Error(), path)
		}
	}
	data.mtx = &versionFileMutex

	return &data, utils.CombineErrors(data.validate()...)
}
