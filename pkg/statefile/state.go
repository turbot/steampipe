package statefile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/migrate"
)

const StateStructVersion = 20220411

// State is a struct containing installation state
type State struct {
	LastCheck      string `json:"last_checked"`    // an RFC3339 encoded time stamp
	InstallationID string `json:"installation_id"` // a UUIDv4 string
	StructVersion  int64  `json:"struct_version"`

	// legacy properties included for backwards compatibility with v0.13
	LegacyLastCheck      string `json:"lastChecked"`
	LegacyInstallationID string `json:"installationId"`
}

func newState() State {
	return State{
		InstallationID: newInstallationID(),
		StructVersion:  StateStructVersion,
	}
}

func LoadState() (State, error) {
	currentState := newState()
	if !files.FileExists(filepaths.StateFilePath()) {
		return currentState, nil
	}

	stateFileContent, err := os.ReadFile(filepaths.StateFilePath())
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
func (s *State) Save() error {
	// set the struct version
	s.StructVersion = StateStructVersion

	s.LastCheck = nowTimeString()
	s.LegacyLastCheck = nowTimeString()
	// maintain the legacy properties for backward compatibility
	s.MaintainLegacy()
	// ensure internal dirs exists
	_ = os.MkdirAll(filepaths.EnsureInternalDir(), os.ModePerm)
	stateFilePath := filepath.Join(filepaths.EnsureInternalDir(), filepaths.StateFileName())
	// if there is an existing file it must be bad/corrupt, so delete it
	_ = os.Remove(stateFilePath)
	// save state file
	file, _ := json.MarshalIndent(s, "", " ")
	return os.WriteFile(stateFilePath, file, 0644)
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (s *State) IsValid() bool {
	return s.StructVersion > 0
}

func (s *State) MigrateFrom() migrate.Migrateable {
	// save the existing property values to the new legacy properties
	s.StructVersion = StateStructVersion
	s.LastCheck = s.LegacyLastCheck
	s.InstallationID = s.LegacyInstallationID

	return s
}

// MaintainLegacy keeps the values of the legacy properties for backward
// compatibility
func (s *State) MaintainLegacy() {
	s.LegacyLastCheck = s.LastCheck
	s.LegacyInstallationID = s.InstallationID
}

func newInstallationID() string {
	return uuid.New().String()
}

func nowTimeString() string {
	return time.Now().Format(time.RFC3339)
}
