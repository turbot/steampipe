package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
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

func pidExists(pid int) (bool, error) {
	if pid <= 0 {
		return false, fmt.Errorf("invalid pid %v", pid)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if err.Error() == "os: process already finished" {
		return false, nil
	}
	errno, ok := err.(syscall.Errno)
	if !ok {
		return false, err
	}
	switch errno {
	case syscall.ESRCH:
		return false, nil
	case syscall.EPERM:
		return true, nil
	}
	return false, err
}
