package db_local

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// GetState checks that the database instance is running and returns its details
func GetState() (*RunningDBInstanceInfo, error) {
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

	pidExists := utils.PidExists(info.Pid)
	if !pidExists {
		log.Printf("[TRACE] GetState - pid %v does not exist\n", info.Pid)
		// nothing to do here
		os.Remove(filepaths.RunningInfoFilePath())
		return nil, nil
	}

	return info, nil
}

// errorIfUnknownService returns an error if it can find a `postmaster.pid` in the `INSTALL_DIR`
// and the PID recorded in the found `postmaster.pid` is running - nil otherwise.
//
// This is because, this function is called when we cannot find the steampipe service state file.
//
// No steampipe state file indicates that the service is not running, so, if the service
// is running without us knowing about it, then it's an irrecoverable state
func errorIfUnknownService() error {
	// no postmaster.pid, we are good
	if !filehelpers.FileExists(filepaths.GetPostmasterPidLocation()) {
		return nil
	}

	// read the content of the postmaster.pid file
	fileContent, err := os.ReadFile(filepaths.GetPostmasterPidLocation())
	if err != nil {
		return err
	}

	// the first line contains the PID
	lines := strings.FieldsFunc(string(fileContent), func(r rune) bool {
		return r == '\n'
	})

	// make sure that there's split up content
	if len(lines) == 0 {
		return nil
	}

	// extract it
	pid, err := strconv.ParseInt(lines[0], 10, 64)
	if err != nil {
		return err
	}

	// check if a process with that PID exists
	exists := utils.PidExists(int(pid))
	if exists {
		// if it does, then somehow we don't know about it. Error out
		return fmt.Errorf("service is running in an unknown state [PID: %d] - try killing it with %s", pid, constants.Bold("steampipe service stop --force"))
	}

	// the pid does not exist
	// this can confuse postgres as per https://postgresapp.com/documentation/troubleshooting.html
	// delete it
	os.Remove(filepaths.GetPostmasterPidLocation())

	// this must be a stale file left over by PG. Ignore
	return nil
}
