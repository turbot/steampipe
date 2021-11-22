package db_local

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
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

func (r *RunningDBInstanceInfo) Save() error {
	content, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(constants.RunningInfoFilePath(), content, 0644)
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

	if !helpers.FileExists(constants.RunningInfoFilePath()) {
		return nil, nil
	}

	fileContent, err := ioutil.ReadFile(constants.RunningInfoFilePath())
	if err != nil {
		return nil, err
	}
	var info = new(RunningDBInstanceInfo)
	err = json.Unmarshal(fileContent, info)
	if err != nil {
		log.Printf("[TRACE] failed to unmarshal database state file %s: %s\n", constants.RunningInfoFilePath(), err.Error())
		return nil, nil
	}
	return info, nil
}

func removeRunningInstanceInfo() error {
	return os.Remove(constants.RunningInfoFilePath())
}
