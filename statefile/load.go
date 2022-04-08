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

const legacyStateFileName = "update-check.json"
const updateStateFileName = "update_check.json"

// State :: the state of the installation
type LegacyState struct {
	LastCheck      string `json:"lastChecked"`    // an RFC3339 encoded time stamp
	InstallationID string `json:"installationId"` // a UUIDv4 string
}

type State struct {
	LastCheck      string `json:"last_checked"`    // an RFC3339 encoded time stamp
	InstallationID string `json:"installation_id"` // a UUIDv4 string
	SchemaVersion  string `json:"schema_version"`
}

func (s State) IsValid() bool {
	return len(s.SchemaVersion) > 0
}

func (s State) MigrateFrom(oldI interface{}) migrate.Migrateable {
	old := oldI.(LegacyState)
	s.SchemaVersion = "20220407"
	s.LastCheck = old.LastCheck
	s.InstallationID = old.InstallationID

	return s
}

func (s State) WriteOut() error {
	// ensure internal dirs exists
	if err := os.MkdirAll(filepaths.EnsureInternalDir(), os.ModePerm); err != nil {
		return err
	}
	stateFilePath := filepath.Join(filepaths.EnsureInternalDir(), updateStateFileName)
	// if there is an existing file it must be bad/corrupt, so delete it
	_ = os.Remove(stateFilePath)
	// save state file
	file, _ := json.MarshalIndent(s, "", " ")
	return os.WriteFile(stateFilePath, file, 0644)
}

func LegacyStateFilePath() string {
	return filepath.Join(filepaths.EnsureInternalDir(), legacyStateFileName)
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
