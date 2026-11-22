package db_local

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	psutils "github.com/shirou/gopsutil/process"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	sputils "github.com/turbot/steampipe/v2/pkg/utils"
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

	// pipe-fittings PidExists only asks whether the pid is in the process table.
	// That is true for a zombie, and for a different process that reused the pid.
	running, err := pidIsLiveSteampipeDatabase(info.Pid)
	if err != nil {
		return nil, err
	}
	if !running {
		log.Printf("[TRACE] GetState - pid %v is not a running steampipe database\n", info.Pid)
		// nothing to do here
		os.Remove(filepaths.RunningInfoFilePath())
		return nil, nil
	}

	return info, nil
}

// pidIsLiveSteampipeDatabase reports whether pid should be treated as a running
// steampipe database. It is false only when the pid is known not to be one: the
// process has exited (including a zombie still in the process table), or its
// readable command line belongs to some other program that reused the pid.
// A live process whose command line cannot be confirmed either way is treated
// as running, so its state files are kept.
func pidIsLiveSteampipeDatabase(pid int) (bool, error) {
	process, err := sputils.FindProcess(pid)
	if err != nil || process == nil {
		return false, err
	}
	ctx := context.Background()
	// reuse the service scan, including its retry for a command line that is still empty
	listProcess := func(context.Context) ([]*psutils.Process, error) {
		return []*psutils.Process{process}, nil
	}
	instances, err := findSteampipePostgresInstances(ctx, listProcess)
	if err != nil {
		return false, err
	}
	if len(instances) > 0 {
		return true, nil
	}
	cmdline, err := readProcessCmdline(ctx, process)
	// alive but unreadable, or still empty: cannot say it is not the database.
	// The read error is not returned. Doing so would drop the state of a live
	// service whose command line this process cannot read.
	unreadable := err != nil || len(cmdline) == 0
	if unreadable {
		return true, nil
	}
	// the scan skips a process it could not read. A later read uses the same
	// matcher, including a postgres binary whose path was split at a space.
	return isSteampipePostgresProcess(ctx, cmdline), nil
}

// errorIfUnknownService returns an error if it can find a `postmaster.pid` in the `INSTALL_DIR`
// and the PID recorded there is a running steampipe database - nil otherwise.
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

	// the same check as the service state file: a zombie, or another process
	// that reused this pid, is not an unknown running database
	exists, err := pidIsLiveSteampipeDatabase(int(pid))
	if err != nil {
		return err
	}
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
