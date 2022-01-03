package db_local

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/utils"
)

// RunningDBInstanceInfo contains data about the running process and it's credentials
type RunningDBInstanceInfo struct {
	Pid        int
	Port       int
	Listen     []string
	ListenType StartListenType
	Invoker    constants.Invoker
	Password   string
	User       string
	Database   string
}

func newRunningDBInstanceInfo(cmd *exec.Cmd, port int, databaseName string, password string, listen StartListenType, invoker constants.Invoker) *RunningDBInstanceInfo {
	dbState := new(RunningDBInstanceInfo)
	dbState.Pid = cmd.Process.Pid
	dbState.Port = port
	dbState.User = constants.DatabaseUser
	dbState.Password = password
	dbState.Database = databaseName
	dbState.ListenType = listen
	dbState.Invoker = invoker
	dbState.Listen = constants.DatabaseListenAddresses

	if listen == ListenTypeNetwork {
		addrs, _ := localAddresses()
		dbState.Listen = append(dbState.Listen, addrs...)
	}
	return dbState
}
func (r *RunningDBInstanceInfo) Save() error {
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
	jsonEncoder.Encode(r)
	r.Password = p
	return writeBuffer.String()
}

func loadRunningInstanceInfo() (*RunningDBInstanceInfo, error) {
	utils.LogTime("db.loadRunningInstanceInfo start")
	defer utils.LogTime("db.loadRunningInstanceInfo end")

	if !helpers.FileExists(filepaths.RunningInfoFilePath()) {
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
