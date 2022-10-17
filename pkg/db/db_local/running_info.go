package db_local

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
)

const RunningDBStructVersion = 20220411

// RunningDBInstanceInfo contains data about the running process and it's credentials
type RunningDBInstanceInfo struct {
	Pid           int               `json:"pid"`
	Port          int               `json:"port"`
	Listen        []string          `json:"listen"`
	ListenType    StartListenType   `json:"listen_type"`
	Invoker       constants.Invoker `json:"invoker"`
	Password      string            `json:"password"`
	User          string            `json:"user"`
	Database      string            `json:"database"`
	StructVersion int64             `json:"struct_version"`
}

func newRunningDBInstanceInfo(cmd *exec.Cmd, port int, databaseName string, password string, listen StartListenType, invoker constants.Invoker) *RunningDBInstanceInfo {
	dbState := &RunningDBInstanceInfo{
		Pid:           cmd.Process.Pid,
		Port:          port,
		User:          constants.DatabaseUser,
		Password:      password,
		Database:      databaseName,
		ListenType:    listen,
		Invoker:       invoker,
		Listen:        constants.DatabaseListenAddresses,
		StructVersion: RunningDBStructVersion,
	}

	if listen == ListenTypeNetwork {
		addrs, _ := utils.LocalAddresses()
		dbState.Listen = append(dbState.Listen, addrs...)
	}

	return dbState
}

func (r *RunningDBInstanceInfo) Save() error {
	// set struct version
	r.StructVersion = RunningDBStructVersion

	content, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepaths.RunningInfoFilePath(), content, 0644)
}

func (r *RunningDBInstanceInfo) String() string {
	writeBuffer := bytes.NewBufferString("")
	jsonEncoder := json.NewEncoder(writeBuffer)

	// redact the password from the string, so that it doesn't get printed
	// this should not affect the state file, since we use a json.Marshal there
	p := r.Password
	r.Password = "XXXX-XXXX-XXXX"

	jsonEncoder.SetIndent("", "")
	err := jsonEncoder.Encode(r)
	if err != nil {
		log.Printf("[TRACE] Encode failed: %v\n", err)
	}
	r.Password = p
	return writeBuffer.String()
}

func loadRunningInstanceInfo() (*RunningDBInstanceInfo, error) {
	utils.LogTime("db.loadRunningInstanceInfo start")
	defer utils.LogTime("db.loadRunningInstanceInfo end")

	if !filehelpers.FileExists(filepaths.RunningInfoFilePath()) {
		return nil, nil
	}

	fileContent, err := os.ReadFile(filepaths.RunningInfoFilePath())
	if err != nil {
		return nil, err
	}
	var info = new(RunningDBInstanceInfo)
	err = json.Unmarshal(fileContent, info)
	if err != nil {
		log.Printf("[TRACE] failed to unmarshal database state file %s: %s\n", filepaths.RunningInfoFilePath(), err.Error())
		return nil, nil
	}
	return info, nil
}

func removeRunningInstanceInfo() error {
	return os.Remove(filepaths.RunningInfoFilePath())
}
