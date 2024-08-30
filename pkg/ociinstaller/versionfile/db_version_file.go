package versionfile

import (
	"encoding/json"
	versionfile2 "github.com/turbot/pipe-fittings/ociinstaller/versionfile"
	"log"
	"os"
	"time"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
)

const DatabaseStructVersion = 20220411

type DatabaseVersionFile struct {
	FdwExtension  versionfile2.InstalledVersion `json:"fdw_extension"`
	EmbeddedDB    versionfile2.InstalledVersion `json:"embedded_db"`
	StructVersion int64                         `json:"struct_version"`
}

func NewDBVersionFile() *DatabaseVersionFile {
	return &DatabaseVersionFile{
		FdwExtension:  versionfile2.InstalledVersion{},
		EmbeddedDB:    versionfile2.InstalledVersion{},
		StructVersion: DatabaseStructVersion,
	}
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (s DatabaseVersionFile) IsValid() bool {
	return s.StructVersion > 0
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
	if data.FdwExtension == (versionfile2.InstalledVersion{}) {
		data.FdwExtension = versionfile2.InstalledVersion{}
	}

	if data.EmbeddedDB == (versionfile2.InstalledVersion{}) {
		data.EmbeddedDB = versionfile2.InstalledVersion{}
	}

	return &data, nil
}

// Save writes the config
func (f *DatabaseVersionFile) Save() error {
	// set the struct version
	f.StructVersion = DatabaseStructVersion

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
