package installationstate

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

const StateStructVersion = 20220411

// stateMutex protects concurrent writes to the state file
var stateMutex sync.Mutex

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
	// protect concurrent writes to the state file
	stateMutex.Lock()
	defer stateMutex.Unlock()

	// set the struct version
	s.StructVersion = StateStructVersion

	s.LastCheck = nowTimeString()
	// ensure internal dirs exists
	_ = os.MkdirAll(filepaths.EnsureInternalDir(), os.ModePerm)
	stateFilePath := filepath.Join(filepaths.EnsureInternalDir(), filepaths.StateFileName())
	file, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}

	// write to a temp file, then atomically rename it into place, so a
	// concurrent Load() never sees the file disappear or contain partial
	// JSON in the window between its existence check and its read (the
	// mutex above serializes writers within this process; it does not
	// protect against two separate steampipe processes saving at once,
	// matching pluginmanager/state.go)
	tempFile := stateFilePath + ".tmp"
	if err := os.WriteFile(tempFile, file, 0644); err != nil {
		return err
	}
	return os.Rename(tempFile, stateFilePath)
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
