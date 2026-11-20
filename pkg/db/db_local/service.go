package db_local

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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

// pidIsLiveSteampipeDatabase reports whether pid is a running steampipe
// database process. A process that has exited, including one still in the
// process table as a zombie, is not. Neither is a different process that has
// reused the pid recorded for the database.
func pidIsLiveSteampipeDatabase(pid int) (bool, error) {
	process, err := sputils.FindProcess(pid)
	if err != nil || process == nil {
		return false, err
	}
	return processIsSteampipeDatabase(context.Background(), process)
}

// processIsSteampipeDatabase applies the same command-line identity check as
// the service scan. A process that has only just started can report an empty
// command line for a short time, so that case is read again.
func processIsSteampipeDatabase(ctx context.Context, p *psutils.Process) (bool, error) {
	attempts := 1
	if createdMs, err := p.CreateTimeWithContext(ctx); err == nil &&
		time.Since(time.UnixMilli(createdMs)) < recentlyStartedWindow {
		attempts = emptyCmdlineRetries + 1
	}
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(emptyCmdlineRetryDelay):
			}
		}
		cmdline, err := readProcessCmdline(ctx, p)
		if err != nil {
			return false, nil
		}
		if len(cmdline) == 0 {
			continue
		}
		return isSteampipePostgresProcess(ctx, cmdline), nil
	}
	return false, nil
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
