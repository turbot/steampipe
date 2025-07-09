package versionfile

import (
	"encoding/json"
	"log"
	"os"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/versionfile"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

const DatabaseStructVersion = 20220411

type DatabaseVersionFile struct {
	FdwExtension  versionfile.InstalledVersion `json:"fdw_extension"`
	EmbeddedDB    versionfile.InstalledVersion `json:"embedded_db"`
	StructVersion int64                        `json:"struct_version"`
}

func NewDBVersionFile() *DatabaseVersionFile {
	return &DatabaseVersionFile{
		FdwExtension:  versionfile.InstalledVersion{},
		EmbeddedDB:    versionfile.InstalledVersion{},
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
