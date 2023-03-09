package versionfile

import (
	"encoding/json"
	"log"
	"os"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
)

// LegacyCompositeVersionFile is the composite version file used before v0.7.0, which contained
// both db and plugin properties, now split into two different files
type LegacyCompositeVersionFile struct {
	Plugins      map[string]*InstalledVersion `json:"plugins"`
	FdwExtension InstalledVersion             `json:"fdwExtension"`
	EmbeddedDB   InstalledVersion             `json:"embeddedDB"`
}

// LoadLegacyVersionFile loads the legacy version file, or returns nil if it does not exist
func LoadLegacyVersionFile() (*LegacyCompositeVersionFile, error) {
	versionFilePath := filepaths.LegacyVersionFilePath()
	if filehelpers.FileExists(versionFilePath) {
		return readLegacyVersionFile(versionFilePath)
	}
	return nil, nil
}

func readLegacyVersionFile(path string) (*LegacyCompositeVersionFile, error) {
	file, _ := os.ReadFile(path)

	var data LegacyCompositeVersionFile

	if err := json.Unmarshal([]byte(file), &data); err != nil {
		log.Println("[ERROR]", "Error while reading version file", err)
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

func migrateVersionFiles() (*PluginVersionFile, *DatabaseVersionFile, error) {
	legacyVersionFile, err := LoadLegacyVersionFile()
	if err != nil {
		return nil, nil, err
	}
	if legacyVersionFile == nil {
		return nil, nil, nil
	}

	log.Printf("[TRACE] migrating version file from '%s' to '%s' and '%s'\n",
		filepaths.LegacyVersionFilePath(),
		filepaths.DatabaseVersionFilePath(),
		filepaths.PluginVersionFilePath())

	pluginVersionFile := pluginVersionFileFromLegacy(legacyVersionFile)
	databaseVersionFile := databaseVersionFileFromLegacy(legacyVersionFile)

	// save the new files and remove the old one
	if err := pluginVersionFile.Save(); err != nil {
		return nil, nil, err
	}
	if err := databaseVersionFile.Save(); err != nil {
		// delete the plugin version file which we have already saved
		pluginVersionFile.delete()
		return nil, nil, err
	}
	legacyVersionFile.delete()
	return pluginVersionFile, databaseVersionFile, nil
}

// delete the file on disk if it exists
func (f *LegacyCompositeVersionFile) delete() {
	versionFilePath := filepaths.LegacyVersionFilePath()
	if filehelpers.FileExists(versionFilePath) {
		os.Remove(versionFilePath)
	}
}
