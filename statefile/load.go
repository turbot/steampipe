package statefile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/turbot/steampipe/constants"
)

const updateStateFileName = "update-check.json"

// State :: the state of the installation
type State struct {
	LastCheck      string `json:"lastChecked"`    // an RFC3339 encoded time stamp
	InstallationID string `json:"installationId"` // a UUIDv4 string
}

func LoadState() (*State, error) {
	stateFilePath := filepath.Join(constants.InternalDir(), updateStateFileName)
	// get the state file
	_, err := os.Stat(stateFilePath)
	if err != nil {
		return nil, err
	}

	stateFileContent, err := ioutil.ReadFile(stateFilePath)
	if err != nil {
		fmt.Println("Could not read update state file")
		return nil, err
	}

	currentState := State{}

	err = json.Unmarshal(stateFileContent, &currentState)
	if err != nil {
		fmt.Println("Could not parse update state file")
		return nil, err
	}

	return &currentState, nil
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
