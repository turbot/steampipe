package local_db

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// GetStatus :: check that the db instance is running and returns it's details
func GetStatus() (*RunningDBInstanceInfo, error) {
	utils.LogTime("db.GetStatus start")
	defer utils.LogTime("db.GetStatus end")

	info, err := loadRunningInstanceInfo()
	if err != nil {
		return nil, err
	}

	if info == nil {
		log.Println("[TRACE] GetRunStatus - loadRunningInstanceInfo returned nil ")
		// we do not have a info file
		return nil, errorIfUnknownService()
	}

	pidExists, err := PidExists(info.Pid)
	if err != nil {
		return nil, err
	}
	if !pidExists {
		log.Printf("[TRACE] GetRunStatus - pid %v does not exist\n", info.Pid)
		// nothing to do here
		os.Remove(runningInfoFilePath())
		return nil, nil
	}

	return info, nil
}

// errorIfUnknownService errors if it can find a `postmaster.pid` in the `INSTALL_DIR`
// and the PID recorded in the found `postmaster.pid` is running
func errorIfUnknownService() error {
	// no postmaster.pid, we are good
	if !helpers.FileExists(getPostmasterPidLocation()) {
		return nil
	}

	// read the content of the postmaster.pid file
	fileContent, err := ioutil.ReadFile(getPostmasterPidLocation())
	if err != nil {
		return err
	}

	// the first line contains the PID
	lines := strings.FieldsFunc(string(fileContent), func(r rune) bool {
		return r == '\n'
	})
	// extract it
	pid, err := strconv.ParseInt(lines[0], 10, 64)
	if err != nil {
		return err
	}

	// check if a process with that PID exists
	exists, err := PidExists(int(pid))
	if err != nil {
		return err
	}
	if exists {
		// if it does, then somehow we don't know about it. Error out
		return fmt.Errorf("service is running in an unknown state [%d] - try killing it with %s", pid, constants.Bold("steampipe service stop --force"))
	}

	// this must be a stale file left over by PG. Ignore
	return nil
}
