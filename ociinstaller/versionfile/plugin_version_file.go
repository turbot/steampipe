package versionfile

import (
	"encoding/json"
	"log"
	"os"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/filepaths"
)

type PluginVersionFile struct {
	Plugins map[string]*InstalledVersion `json:"plugins"`
}

func NewPluginVersionFile() *PluginVersionFile {
	return &PluginVersionFile{
		Plugins: map[string]*InstalledVersion{},
	}
}

func pluginVersionFileFromLegacy(legacyFile *LegacyVersionFile) *PluginVersionFile {
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
