package statefile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/migrate"
)

const StructVersion = 20220411

// LegacyState is a struct used to migrate the
// State to serialize with snake case property names
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

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (s State) IsValid() bool {
	return s.StructVersion > 0
}

func (s *State) MigrateFrom(prev interface{}) migrate.Migrateable {
	legacyState := prev.(LegacyState)
	s.StructVersion = StructVersion
	s.LastCheck = legacyState.LastCheck
	s.InstallationID = legacyState.InstallationID

	return s
}

func LoadState() (State, error) {
	currentState := createState()

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

func createState() State {
	// start new current state
	return State{
		InstallationID: newInstallationID(), // a new ID
	}
}

func newInstallationID() string {
	return uuid.New().String()
}

func nowTimeString() string {
	return time.Now().Format(time.RFC3339)
}
