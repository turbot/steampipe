package versionfile

import (
	"encoding/json"
	"log"
	"os"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/migrate"
)

const PluginStructVersion = 20220411

// LegacyPluginVersionFile is a struct used to migrate the
// PluginVersionFile to serialize with snake case property names(migrated in v0.14.0)
type LegacyPluginVersionFile struct {
	Plugins map[string]*LegacyInstalledVersion `json:"plugins"`
}

type PluginVersionFile struct {
	Plugins       map[string]*InstalledVersion `json:"plugins"`
	StructVersion int64                        `json:"struct_version"`
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (s PluginVersionFile) IsValid() bool {
	return s.StructVersion > 0
}

func (s *PluginVersionFile) MigrateFrom(prev interface{}) migrate.Migrateable {
	legacyState := prev.(LegacyPluginVersionFile)
	s.StructVersion = PluginStructVersion
	s.Plugins = make(map[string]*InstalledVersion, len(legacyState.Plugins))
	for p, i := range legacyState.Plugins {
		s.Plugins[p] = &InstalledVersion{
			Name:            i.Name,
			Version:         i.Version,
			ImageDigest:     i.ImageDigest,
			InstalledFrom:   i.InstalledFrom,
			LastCheckedDate: i.LastCheckedDate,
			InstallDate:     i.InstallDate,
		}
	}
	return s
}

func NewPluginVersionFile() *PluginVersionFile {
	return &PluginVersionFile{
		Plugins: map[string]*InstalledVersion{},
	}
}

func pluginVersionFileFromLegacy(legacyFile *LegacyCompositeVersionFile) *PluginVersionFile {
	return &PluginVersionFile{
		Plugins: legacyFile.Plugins,
	}
}

// LoadPluginVersionFile migrates from the old version file format if necessary and loads the plugin version data
func LoadPluginVersionFile() (*PluginVersionFile, error) {
	// first, see if a migration is necessary - if so, it will return the version data to us
	migratedVersionFile, err := MigratePluginVersionFile()
	if err != nil {
		return nil, err
	}
	if migratedVersionFile != nil {
		log.Println("[TRACE] using migrated plugin version file")
		return migratedVersionFile, nil
	}

	versionFilePath := filepaths.PluginVersionFilePath()
	if helpers.FileExists(versionFilePath) {
		return readPluginVersionFile(versionFilePath)
	}
	return NewPluginVersionFile(), nil
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

	return &data, nil
}

// Save writes the config file to disk
func (f *PluginVersionFile) Save() error {
	versionFilePath := filepaths.PluginVersionFilePath()
	return f.write(versionFilePath)
}

func (f *PluginVersionFile) write(path string) error {
	versionFileJSON, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		log.Println("[ERROR]", "Error while writing version file", err)
		return err
	}
	return os.WriteFile(path, versionFileJSON, 0644)
}

// delete the file on disk if it exists
func (f *PluginVersionFile) delete() {
	versionFilePath := filepaths.PluginVersionFilePath()
	if helpers.FileExists(versionFilePath) {
		os.Remove(versionFilePath)
	}
}
