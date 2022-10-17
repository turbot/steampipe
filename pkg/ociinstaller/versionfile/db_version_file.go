package versionfile

import (
	"encoding/json"
	"log"
	"os"
	"time"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/migrate"
)

const DatabaseStructVersion = 20220411

type DatabaseVersionFile struct {
	FdwExtension  InstalledVersion `json:"fdw_extension"`
	EmbeddedDB    InstalledVersion `json:"embedded_db"`
	StructVersion int64            `json:"struct_version"`

	// legacy properties included for backwards compatibility with v0.13
	LegacyFdwExtension InstalledVersion `json:"fdwExtension"`
	LegacyEmbeddedDB   InstalledVersion `json:"embeddedDB"`
}

func NewDBVersionFile() *DatabaseVersionFile {
	return &DatabaseVersionFile{
		FdwExtension:  InstalledVersion{},
		EmbeddedDB:    InstalledVersion{},
		StructVersion: DatabaseStructVersion,
	}
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (s DatabaseVersionFile) IsValid() bool {
	return s.StructVersion > 0
}

func (s *DatabaseVersionFile) MigrateFrom() migrate.Migrateable {
	s.StructVersion = DatabaseStructVersion
	s.FdwExtension = s.LegacyFdwExtension
	s.FdwExtension.MigrateLegacy()
	s.EmbeddedDB = s.LegacyEmbeddedDB
	s.EmbeddedDB.MigrateLegacy()

	return s
}

func databaseVersionFileFromLegacy(legacyFile *LegacyCompositeVersionFile) *DatabaseVersionFile {
	return &DatabaseVersionFile{
		FdwExtension: legacyFile.FdwExtension,
		EmbeddedDB:   legacyFile.EmbeddedDB,
	}
}

// LoadDatabaseVersionFile migrates from the old version file format if necessary and loads the database version data
func LoadDatabaseVersionFile() (*DatabaseVersionFile, error) {
	versionFilePath := filepaths.DatabaseVersionFilePath()
	if filehelpers.FileExists(versionFilePath) {
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
	// set the struct version
	f.StructVersion = DatabaseStructVersion
	// maintain the legacy properties for backward compatibility
	f.LegacyFdwExtension = f.FdwExtension
	f.LegacyFdwExtension.MaintainLegacy()
	f.LegacyEmbeddedDB = f.EmbeddedDB
	f.LegacyEmbeddedDB.MaintainLegacy()

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
