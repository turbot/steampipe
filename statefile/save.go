package statefile

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/constants"
)

func (s *State) Save() error {
	s.LastCheck = nowTimeString()
	// ensure internal dirs exists
	_ = os.MkdirAll(constants.InternalDir(), os.ModePerm)
	stateFilePath := filepath.Join(constants.InternalDir(), updateStateFileName)
	// if there is an existing file it must be bad/corrupt, so delete it
	_ = os.Remove(stateFilePath)
	// save state file
	file, _ := json.MarshalIndent(s, "", " ")
	return ioutil.WriteFile(stateFilePath, file, 0644)
}
