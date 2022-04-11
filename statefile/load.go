package statefile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/migrate"
)

const updateStateFileName = "update_check.json"

// LegacyState is the legacy state struct, which was used in the legacy
// state file
type LegacyState struct {
	LastCheck      string `json:"lastChecked"`    // an RFC3339 encoded time stamp
	InstallationID string `json:"installationId"` // a UUIDv4 string
}

// State :: the state of the installation
type State struct {
	LastCheck      string `json:"last_checked"`    // an RFC3339 encoded time stamp
	InstallationID string `json:"installation_id"` // a UUIDv4 string
	SchemaVersion  string `json:"schema_version"`
}

func (s State) IsValid() bool {
	return len(s.SchemaVersion) > 0
}

func (s *State) MigrateFrom(legacyState interface{}) migrate.Migrateable {
	old := legacyState.(LegacyState)
	s.SchemaVersion = constants.SchemaVersion
	s.LastCheck = old.LastCheck
	s.InstallationID = old.InstallationID

	return s
}

func LoadState() (State, error) {
	currentState := createState()

	stateFilePath := filepath.Join(filepaths.EnsureInternalDir(), updateStateFileName)
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
