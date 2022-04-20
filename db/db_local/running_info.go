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
	"github.com/turbot/steampipe/migrate"
	"github.com/turbot/steampipe/utils"
)

const RunningDBStructVersion = 20220411

// LegacyRunningDBInstanceInfo is a struct used to migrate the
// RunningDBInstanceInfo to serialize with snake case property names(migrated in v0.14.0)
type LegacyRunningDBInstanceInfo struct {
	Pid        int
	Port       int
	Listen     []string
	ListenType StartListenType
	Invoker    constants.Invoker
	Password   string
	User       string
	Database   string
}

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

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (r RunningDBInstanceInfo) IsValid() bool {
	return r.StructVersion > 0
}

func (r *RunningDBInstanceInfo) MigrateFrom(prev interface{}) migrate.Migrateable {
	legacyState := prev.(LegacyRunningDBInstanceInfo)
	r.StructVersion = RunningDBStructVersion
	r.Pid = legacyState.Pid
	r.Port = legacyState.Port
	r.Listen = legacyState.Listen
	r.ListenType = legacyState.ListenType
	r.Invoker = legacyState.Invoker
	r.Password = legacyState.Password
	r.User = legacyState.User
	r.Database = legacyState.Database

	return r
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

func (r *RunningDBInstanceInfo) Save() ([]byte, error) {
	// set struct version
	r.StructVersion = RunningDBStructVersion

	content, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, err
	}
	return content, os.WriteFile(filepaths.RunningInfoFilePath(), content, 0644)
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
