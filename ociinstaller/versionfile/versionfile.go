package versionfile

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/constants"
)

const (
	pluginVersionFileName = "versions.json"
	dbVersionFileName     = "versions.json"
)

type PluginVersionFile struct {
	Plugins map[string]*InstalledVersion `json:"plugins"`
}

type DBVersionFile struct {
	FdwExtension InstalledVersion `json:"fdwExtension"`
	EmbeddedDB   InstalledVersion `json:"embeddedDB"`
}

func NewPluginVersionFile() *PluginVersionFile {
	return &PluginVersionFile{
		Plugins: map[string]*InstalledVersion{},
	}
}

func NewDBVersionFile() *DBVersionFile {
	return &DBVersionFile{
		FdwExtension: InstalledVersion{},
		EmbeddedDB:   InstalledVersion{},
	}
}

type InstalledVersion struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	ImageDigest     string `json:"imageDigest"`
	InstalledFrom   string `json:"installedFrom"`
	LastCheckedDate string `json:"lastCheckedDate"`
	InstallDate     string `json:"installDate"`
}

// LoadForPlugin :: load the config file for plugin from the standard path, or return an empty one
func LoadForPlugin() (*PluginVersionFile, error) {
	versionFilePath := versionFileLocation()
	if fileExists(versionFilePath) {
		return readForPlugin(versionFilePath)
	}
	return NewPluginVersionFile(), nil
}

// LoadForDB :: load the config file for DB from the standard path, or return an empty one
func LoadForDB() (*DBVersionFile, error) {
	versionFilePath := dbVersionFileLocation()
	if fileExists(versionFilePath) {
		return readforDB(versionFilePath)
	}
	return NewDBVersionFile(), nil
}

// SaveForPlugin :: Save the config file from the standard path, or return an empty one
func (f *PluginVersionFile) SaveForPlugin() error {
	versionFilePath := versionFileLocation()
	return f.writeForPlugin(versionFilePath)
}

// SaveForDB :: Save the config file from the standard path, or return an empty one
func (f *DBVersionFile) SaveForDB() error {
	versionFilePath := versionFileLocation()
	return f.writeForDB(versionFilePath)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (f *PluginVersionFile) writeForPlugin(path string) error {
	versionFileJSON, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		log.Println("[ERROR]", "Error while writing version file", err)
		return err
	}
	return ioutil.WriteFile(path, versionFileJSON, 0644)
}

func (f *DBVersionFile) writeForDB(path string) error {
	versionFileJSON, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		log.Println("[ERROR]", "Error while writing version file", err)
		return err
	}
	return ioutil.WriteFile(path, versionFileJSON, 0644)
}

func readForPlugin(path string) (*PluginVersionFile, error) {
	file, _ := ioutil.ReadFile(path)

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

func readforDB(path string) (*DBVersionFile, error) {
	file, _ := ioutil.ReadFile(path)
	var data DBVersionFile
	if err := json.Unmarshal([]byte(file), &data); err != nil {
		log.Println("[ERROR]", "Error while reading DB version file", err)
		return nil, err
	}
	if data.FdwExtension == (InstalledVersion{}) {
		data.FdwExtension = InstalledVersion{}
	}

	if data.EmbeddedDB == (InstalledVersion{}) {
		data.EmbeddedDB = InstalledVersion{}
	}

	return &data, nil
}

// ex: $CONFIG_DIR/plugins/hub.steampipe.io/turbot/aws/1.1.2/steampipe-plugin-aws
func versionFileLocation() string {
	path := filepath.Join(constants.PluginDir(), pluginVersionFileName)
	return path
}

// ex: $CONFIG_DIR/plugins/hub.steampipe.io/turbot/aws/1.1.2/steampipe-plugin-aws
func dbVersionFileLocation() string {
	path := filepath.Join(constants.DatabaseDir(), pluginVersionFileName)
	return path
}

func oldVersionFileLocation() string {
	path := filepath.Join(constants.InternalDir(), pluginVersionFileName)
	return path
}

// FormatTime :: format time as RFC3339 in UTC
func FormatTime(localTime time.Time) string {
	loc, _ := time.LoadLocation("UTC")
	utcTime := localTime.In(loc)
	return (utcTime.Format(time.RFC3339))
}
