package task

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const minimumMinutesBetweenChecks = 1440 // 1 day
const updateStateFileName = "update-check.json"

type Runner struct {
	currentState state
	shouldRun    bool
}

func NewRunner() *Runner {
	r := new(Runner)
	r.loadState()
	r.shouldRun = r.getShouldRun()
	return r
}

func (r *Runner) Run() {
	if r.shouldRun {
		checkVersion(r.currentState.InstallationID)
		// remove log files older than 7 days
		db.TrimLogs()
		// update last check time
		r.updateState()
	}
}

func (r *Runner) loadState() {
	stateFilePath := filepath.Join(constants.InternalDir(), updateStateFileName)
	// get the state file
	_, err := os.Stat(stateFilePath)
	if err != nil {
		r.currentState = r.createState()
		return
		// create folder structure if not there
	}

	stateFileContent, err := ioutil.ReadFile(stateFilePath)
	if err != nil {
		fmt.Println("Could not read update state file")
		r.currentState = r.createState()
		return
	}

	err = json.Unmarshal(stateFileContent, &r.currentState)
	if err != nil {
		fmt.Println("Could not parse update state file")
		r.currentState = r.createState()
		return
	}
}

func (r *Runner) getShouldRun() bool {
	now := time.Now()
	if r.currentState.LastCheck == "" {
		return true
	}
	lastCheckedAt, err := time.Parse(time.RFC3339, r.currentState.LastCheck)
	if err != nil {
		return true
	}
	minutesElapsed := now.Sub(lastCheckedAt).Minutes()
	return minutesElapsed > minimumMinutesBetweenChecks
}

func (r *Runner) createState() state {
	// start new current state
	r.currentState = state{
		InstallationID: newInstallationId(), // a new ID
	}
	return r.currentState
}

func (r *Runner) updateState() {
	r.currentState.LastCheck = nowTimeString()
	// ensure internal dirs exists
	_ = os.MkdirAll(constants.InternalDir(), os.ModePerm)
	stateFilePath := filepath.Join(constants.InternalDir(), updateStateFileName)
	// if there is an existing file it must be bad/corrupt, so deleted
	_ = os.Remove(stateFilePath)
	// save state file
	file, _ := json.MarshalIndent(r.currentState, "", " ")
	_ = ioutil.WriteFile(stateFilePath, file, 0644)
}

func newInstallationId() string {
	return uuid.New().String()
}

func nowTimeString() string {
	return time.Now().Format(time.RFC3339)
}
