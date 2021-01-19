package versionfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/constants"
)

const (
	VersionFileName = "versions.json"
)

type VersionFile struct {
	Plugins      map[string]*InstalledVersion `json:"plugins"`
	FdwExtension InstalledVersion             `json:"fdwExtension"`
	EmbeddedDB   InstalledVersion             `json:"embeddedDB"`
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
	return new(VersionFile), nil
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
		fmt.Println("---- error: ", err)
		return err
	}
	return ioutil.WriteFile(path, versionFileJSON, 0644)
}

func read(path string) (*VersionFile, error) {
	file, _ := ioutil.ReadFile(path)

	var data VersionFile

	if err := json.Unmarshal([]byte(file), &data); err != nil {
		fmt.Println("---- error: ", err)
		return nil, err
	}

	return &data, nil
}

// ex: $CONFIG_DIR/plugins/registry.steampipe.io/turbot/aws/1.1.2/steampipe-plugin-aws
func versionFileLocation() string {
	path := filepath.Join(constants.InternalDir(), VersionFileName)
	return path
}

// FormatTime :: format time as RFC3339 in UTC
func FormatTime(localTime time.Time) string {
	loc, _ := time.LoadLocation("UTC")
	utcTime := localTime.In(loc)
	return (utcTime.Format(time.RFC3339))
}
