package db

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// RunningDBInstanceInfo :: contains data about the running process
// and it's credentials
type RunningDBInstanceInfo struct {
	Pid        int
	Port       int
	Listen     []string
	ListenType StartListenType
	Invoker    Invoker
	Password   string
	User       string
	Database   string
}

func saveRunningInstanceInfo(info *RunningDBInstanceInfo) error {
	if content, err := json.Marshal(info); err != nil {
		return err
	} else {
		return ioutil.WriteFile(runningInfoFilePath(), content, 0644)
	}
}

func loadRunningInstanceInfo() (*RunningDBInstanceInfo, error) {
	utils.LogTime("db.loadRunningInstanceInfo start")
	defer utils.LogTime("db.loadRunningInstanceInfo end")

	if !helpers.FileExists(runningInfoFilePath()) {
		return nil, nil
	}

	fileContent, err := ioutil.ReadFile(runningInfoFilePath())
	if err != nil {
		return nil, err
	}
	var info = new(RunningDBInstanceInfo)
	err = json.Unmarshal(fileContent, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func removeRunningInstanceInfo() error {
	return os.Remove(runningInfoFilePath())
}

func runningInfoFilePath() string {
	return filepath.Join(constants.InternalDir(), "steampipe.json")
}
