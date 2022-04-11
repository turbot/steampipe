package versionfile

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/migrate"
)

// LegacyDatabaseVersionFile is the legacy db version file struct, which was used
// in the legacy db state file
type LegacyDatabaseVersionFile struct {
	FdwExtension LegacyInstalledVersion `json:"fdwExtension"`
	EmbeddedDB   LegacyInstalledVersion `json:"embeddedDB"`
}

type DatabaseVersionFile struct {
	FdwExtension  InstalledVersion `json:"fdw_extension"`
	EmbeddedDB    InstalledVersion `json:"embedded_db"`
	SchemaVersion string           `json:"schema_version"`
}

func NewDBVersionFile() *DatabaseVersionFile {
	return &DatabaseVersionFile{
		FdwExtension: InstalledVersion{},
		EmbeddedDB:   InstalledVersion{},
	}
}

func (s DatabaseVersionFile) IsValid() bool {
	return len(s.SchemaVersion) > 0
}

func (s *DatabaseVersionFile) MigrateFrom(legacyState interface{}) migrate.Migrateable {
	old := legacyState.(LegacyDatabaseVersionFile)
	s.SchemaVersion = constants.SchemaVersion
	s.FdwExtension.Name = old.FdwExtension.Name
	s.FdwExtension.Version = old.FdwExtension.Version
	s.FdwExtension.ImageDigest = old.FdwExtension.ImageDigest
	s.FdwExtension.InstalledFrom = old.FdwExtension.InstalledFrom
	s.FdwExtension.LastCheckedDate = old.FdwExtension.LastCheckedDate
	s.FdwExtension.InstallDate = old.FdwExtension.InstallDate

	s.EmbeddedDB.Name = old.EmbeddedDB.Name
	s.EmbeddedDB.Version = old.EmbeddedDB.Version
	s.EmbeddedDB.ImageDigest = old.EmbeddedDB.ImageDigest
	s.EmbeddedDB.InstalledFrom = old.EmbeddedDB.InstalledFrom
	s.EmbeddedDB.LastCheckedDate = old.EmbeddedDB.LastCheckedDate
	s.EmbeddedDB.InstallDate = old.EmbeddedDB.InstallDate

	return s
}

// LegacyInstalledVersion is the legacy db installed version info struct
type LegacyInstalledVersion struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	ImageDigest     string `json:"imageDigest"`
	InstalledFrom   string `json:"installedFrom"`
	LastCheckedDate string `json:"lastCheckedDate"`
	InstallDate     string `json:"installDate"`
}

type InstalledVersion struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	ImageDigest     string `json:"image_digest"`
	InstalledFrom   string `json:"installed_from"`
	LastCheckedDate string `json:"last_checked_date"`
	InstallDate     string `json:"install_date"`
}

func databaseVersionFileFromLegacy(legacyFile *LegacyCompositeVersionFile) *DatabaseVersionFile {
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
	if err := json.Unmarshal(file, &data); err != nil {
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

// FormatTime :: format time as RFC3339 in UTC
func FormatTime(localTime time.Time) string {
	loc, _ := time.LoadLocation("UTC")
	utcTime := localTime.In(loc)
	return (utcTime.Format(time.RFC3339))
}
