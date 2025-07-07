package installationstate

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

const StateStructVersion = 20220411

type InstallationState struct {
	LastCheck      string `json:"last_checked"`    // an RFC3339 encoded time stamp
	InstallationID string `json:"installation_id"` // a UUIDv4 string
	StructVersion  int64  `json:"struct_version"`
}

func newInstallationState() InstallationState {
	return InstallationState{
		InstallationID: newInstallationID(),
		StructVersion:  StateStructVersion,
	}
}

func Load() (InstallationState, error) {
	currentState := newInstallationState()
	if !files.FileExists(filepaths.StateFilePath()) {
		return currentState, nil
	}

	stateFileContent, err := os.ReadFile(filepaths.StateFilePath())
	if err != nil {
		log.Println("[INFO] Could not read update state file")
		return currentState, err
	}

	err = json.Unmarshal(stateFileContent, &currentState)
	if err != nil {
		log.Println("[INFO] Could not parse update state file")
		return currentState, err
	}

	return currentState, nil
}

// Save the state
// NOTE: this updates the last checked time to the current time
func (s *InstallationState) Save() error {
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
	return os.WriteFile(stateFilePath, file, 0644)
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (s *InstallationState) IsValid() bool {
	return s.StructVersion > 0
}

func newInstallationID() string {
	return uuid.New().String()
}

func nowTimeString() string {
	return time.Now().Format(time.RFC3339)
}
