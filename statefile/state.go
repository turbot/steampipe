package statefile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/migrate"

	"github.com/google/uuid"
	"github.com/turbot/steampipe/filepaths"
)

const StateStructVersion = 20220411

// LegacyState is a struct used to migrate the
// State to serialize with snake case property names(migrated in v0.14.0)
type LegacyState struct {
	LastCheck      string `json:"lastChecked"`    // an RFC3339 encoded time stamp
	InstallationID string `json:"installationId"` // a UUIDv4 string
}

// State is a struct containing installation state
type State struct {
	LastCheck      string `json:"last_checked"`    // an RFC3339 encoded time stamp
	InstallationID string `json:"installation_id"` // a UUIDv4 string
	StructVersion  int64  `json:"struct_version"`
}

func newState() State {
	return State{
		InstallationID: newInstallationID(),
		StructVersion:  StateStructVersion,
	}
}

func LoadState() (State, error) {
	currentState := newState()

	stateFilePath := filepath.Join(filepaths.EnsureInternalDir(), filepaths.StateFileName())
	// get the state file
	_, err := os.Stat(stateFilePath)
	if err != nil {
		return currentState, err
	}

	stateFileContent, err := os.ReadFile(stateFilePath)
	if err != nil {
		fmt.Println("Could not read update state file")
		return currentState, err
	}

	err = json.Unmarshal(stateFileContent, &currentState)
	if err != nil {
		fmt.Println("Could not parse update state file")
		return currentState, err
	}

	return currentState, nil
}

// Save the state
// NOTE: this updates the last checked time
func (s *State) Save() ([]byte, error) {
	// set the struct version
	s.StructVersion = StateStructVersion

	s.LastCheck = nowTimeString()
	// ensure internal dirs exists
	_ = os.MkdirAll(filepaths.EnsureInternalDir(), os.ModePerm)
	stateFilePath := filepath.Join(filepaths.EnsureInternalDir(), filepaths.StateFileName())
	// if there is an existing file it must be bad/corrupt, so delete it
	_ = os.Remove(stateFilePath)
	// save state file
	file, _ := json.MarshalIndent(s, "", " ")
	return file, os.WriteFile(stateFilePath, file, 0644)
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (s *State) IsValid() bool {
	return s.StructVersion > 0
}

func (s *State) MigrateFrom(prev interface{}) migrate.Migrateable {
	legacyState := prev.(LegacyState)
	s.StructVersion = StateStructVersion
	s.LastCheck = legacyState.LastCheck
	s.InstallationID = legacyState.InstallationID

	return s
}

func newInstallationID() string {
	return uuid.New().String()
}

func nowTimeString() string {
	return time.Now().Format(time.RFC3339)
}
