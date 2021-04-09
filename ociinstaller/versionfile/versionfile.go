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
	versionFileName = "versions.json"
)

type VersionFile struct {
	Plugins      map[string]*InstalledVersion `json:"plugins"`
	FdwExtension InstalledVersion             `json:"fdwExtension"`
	EmbeddedDB   InstalledVersion             `json:"embeddedDB"`
}

func NewVersionFile() *VersionFile {
	return &VersionFile{
		Plugins: map[string]*InstalledVersion{},
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

// Load :: load the config file from the standard path, or return an empty one
func Load() (*VersionFile, error) {
	versionFilePath := versionFileLocation()
	if fileExists(versionFilePath) {
		return read(versionFilePath)
	}
	return NewVersionFile(), nil
}

// Save :: Save the config file from the standard path, or return an empty one
func (f *VersionFile) Save() error {
	versionFilePath := versionFileLocation()
	return f.write(versionFilePath)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (f *VersionFile) write(path string) error {
	versionFileJSON, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		log.Println("[ERROR]", "Error while writing version file", err)
		return err
	}
	return ioutil.WriteFile(path, versionFileJSON, 0644)
}

func read(path string) (*VersionFile, error) {
	file, _ := ioutil.ReadFile(path)

	var data VersionFile

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

// ex: $CONFIG_DIR/plugins/hub.steampipe.io/turbot/aws/1.1.2/steampipe-plugin-aws
func versionFileLocation() string {
	path := filepath.Join(constants.InternalDir(), versionFileName)
	return path
}

// FormatTime :: format time as RFC3339 in UTC
func FormatTime(localTime time.Time) string {
	loc, _ := time.LoadLocation("UTC")
	utcTime := localTime.In(loc)
	return (utcTime.Format(time.RFC3339))
}
