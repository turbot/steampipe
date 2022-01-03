package versionfile

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/filepaths"
)

const (
	dbVersionFileName = "versions.json"
)

type DatabaseVersionFile struct {
	FdwExtension InstalledVersion `json:"fdwExtension"`
	EmbeddedDB   InstalledVersion `json:"embeddedDB"`
}

func NewDBVersionFile() *DatabaseVersionFile {
	return &DatabaseVersionFile{
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

func databaseVersionFileFromLegacy(legacyFile *LegacyVersionFile) *DatabaseVersionFile {
	return &DatabaseVersionFile{
		FdwExtension: legacyFile.FdwExtension,
		EmbeddedDB:   legacyFile.EmbeddedDB,
	}
}

// LoadDatabaseVersionFile migrates from the old version file format if necessary and loads the database version data
func LoadDatabaseVersionFile() (*DatabaseVersionFile, error) {
	// first, see if a migration is necessary - if so, it will return the version data to us
	migratedVersionFile, err := MigrateDatabaseVersionFile()
	if err != nil {
		return nil, err
	}
	if migratedVersionFile != nil {
		log.Println("[TRACE] using migrated database version file")
		return migratedVersionFile, nil
	}

	versionFilePath := filepaths.DatabaseVersionFilePath()
	if helpers.FileExists(versionFilePath) {
		return readDatabaseVersionFile(versionFilePath)
	}
	return NewDBVersionFile(), nil
}

func readDatabaseVersionFile(path string) (*DatabaseVersionFile, error) {
	file, _ := os.ReadFile(path)
	var data DatabaseVersionFile
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

// Save writes the config
func (f *DatabaseVersionFile) Save() error {
	versionFilePath := filepaths.DatabaseVersionFilePath()
	return f.write(versionFilePath)
}

func (f *DatabaseVersionFile) write(path string) error {
	versionFileJSON, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		log.Println("[ERROR]", "Error while writing version file", err)
		return err
	}
	return os.WriteFile(path, versionFileJSON, 0644)
}

// delete the file on disk if it exists
func (f *DatabaseVersionFile) delete() {
	versionFilePath := filepaths.DatabaseVersionFilePath()
	if helpers.FileExists(versionFilePath) {
		os.Remove(versionFilePath)
	}
}

// FormatTime :: format time as RFC3339 in UTC
func FormatTime(localTime time.Time) string {
	loc, _ := time.LoadLocation("UTC")
	utcTime := localTime.In(loc)
	return (utcTime.Format(time.RFC3339))
}
