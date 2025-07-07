package db_local

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"slices"
	"sort"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	putils "github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

const RunningDBStructVersion = 20220411

// RunningDBInstanceInfo contains data about the running process and it's credentials
type RunningDBInstanceInfo struct {
	Pid int `json:"pid"`
	// store both resolved and user input listen addresses
	// keep the same 'listen' json tag to maintain backward compatibility
	ResolvedListenAddresses []string          `json:"listen"`
	GivenListenAddresses    []string          `json:"raw_listen"`
	Port                    int               `json:"port"`
	Invoker                 constants.Invoker `json:"invoker"`
	Password                string            `json:"password"`
	User                    string            `json:"user"`
	Database                string            `json:"database"`
	StructVersion           int64             `json:"struct_version"`
}

func newRunningDBInstanceInfo(cmd *exec.Cmd, listenAddresses []string, port int, databaseName string, password string, invoker constants.Invoker) *RunningDBInstanceInfo {
	resolvedListenAddresses := getListenAddresses(listenAddresses)

	dbState := &RunningDBInstanceInfo{
		Pid:                     cmd.Process.Pid,
		ResolvedListenAddresses: resolvedListenAddresses,
		GivenListenAddresses:    listenAddresses,
		Port:                    port,
		User:                    constants.DatabaseUser,
		Password:                password,
		Database:                databaseName,
		Invoker:                 invoker,
		StructVersion:           RunningDBStructVersion,
	}

	return dbState
}

func getListenAddresses(listenAddresses []string) []string {
	addresses := []string{}

	if slices.Contains(listenAddresses, "localhost") {
		loopAddrs, err := putils.LocalLoopbackAddresses()
		if err != nil {
			return nil
		}
		addresses = loopAddrs
	}

	if slices.Contains(listenAddresses, "*") {
		// remove the * wildcard, we want to replace that with the actual addresses
		listenAddresses = helpers.RemoveFromStringSlice(listenAddresses, "*")
		loopAddrs, err := putils.LocalLoopbackAddresses()
		if err != nil {
			return nil
		}
		publicAddrs, err := putils.LocalPublicAddresses()
		if err != nil {
			return nil
		}
		addresses = append(loopAddrs, publicAddrs...)
	}

	// now add back the listenAddresses to address arguments where the interface addresses were sent
	addresses = append(addresses, listenAddresses...)
	addresses = helpers.StringSliceDistinct(addresses)

	// sort locals to the top
	sort.SliceStable(addresses, func(i, j int) bool {
		locals := []string{
			"127.0.0.1",
			"::1",
			"localhost",
		}
		return !slices.Contains(locals, addresses[j])
	})

	return addresses
}

func (r *RunningDBInstanceInfo) MatchWithGivenListenAddresses(listenAddresses []string) bool {
	// make a clone of the slices - we don't want to modify the original data in the subsequent sort
	left := slices.Clone(r.GivenListenAddresses)
	right := slices.Clone(listenAddresses)

	// sort both of them
	slices.Sort(left)
	slices.Sort(right)

	return slices.Equal(left, right)
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
	putils.LogTime("db.loadRunningInstanceInfo start")
	defer putils.LogTime("db.loadRunningInstanceInfo end")

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
