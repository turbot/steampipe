package statefile

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/filepaths"
)

// Save the state
// NOTE: this updates the last checked time
func (s *State) Save() error {
	s.LastCheck = nowTimeString()
	// ensure internal dirs exists
	_ = os.MkdirAll(filepaths.EnsureInternalDir(), os.ModePerm)
	stateFilePath := filepath.Join(filepaths.EnsureInternalDir(), updateStateFileName)
	// if there is an existing file it must be bad/corrupt, so delete it
	_ = os.Remove(stateFilePath)
	// save state file
	file, _ := json.MarshalIndent(s, "", " ")
	return os.WriteFile(stateFilePath, file, 0644)
}
