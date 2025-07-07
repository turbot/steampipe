package db_local

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	psutils "github.com/shirou/gopsutil/process"
	putils "github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/constants/runtime"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	"github.com/turbot/steampipe/v2/pkg/pluginmanager"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
	"github.com/turbot/steampipe/v2/pkg/utils"
)

// StopStatus is a pseudoEnum for service stop result
type StopStatus int

const (
	// start from 1 to prevent confusion with int zero-value
	ServiceStopped StopStatus = iota + 1
	ServiceNotRunning
	ServiceStopFailed
	ServiceStopTimedOut
)

// ShutdownService stops the database instance if the given 'invoker' matches
func ShutdownService(ctx context.Context, invoker constants.Invoker) {
	putils.LogTime("db_local.ShutdownService start")
	defer putils.LogTime("db_local.ShutdownService end")

	if error_helpers.IsContextCanceled(ctx) {
		ctx = context.Background()
	}

	status, _ := GetState()

	// if the service is not running or it was invoked by 'steampipe service',
	// then we don't shut it down
	if status == nil || status.Invoker == constants.InvokerService {
		return
	}

	// how many clients are connected
	// under a fresh context
	clientCounts, err := GetClientCount(context.Background())
	// if there are other clients connected
	// and if there's no error
	if err == nil && clientCounts.SteampipeClients > 0 {
		// there are other steampipe clients connected to the database
		// we don't need to stop the service
		// the last one to exit will shutdown the service
		log.Printf("[INFO] ShutdownService not closing database service - %d steampipe %s connected", clientCounts.SteampipeClients, putils.Pluralize("client", clientCounts.SteampipeClients))
		return
	}

	// we can shut down the database
	stopStatus, err := StopServices(ctx, false, invoker)
	if err != nil {
		error_helpers.ShowError(ctx, err)
	}
	if stopStatus == ServiceStopped {
		return
	}

	// shutdown failed - try to force stop
	_, err = StopServices(ctx, true, invoker)
	if err != nil {
		error_helpers.ShowError(ctx, err)
	}

}

type ClientCount struct {
	SteampipeClients     int
	PluginManagerClients int
	TotalClients         int
}

// GetClientCount returns the number of connections to the service from anyone other than
// _this_execution_ of steampipe
//
// We assume that any connections from this execution will eventually be closed
// - if there are any other external connections, we cannot shut down the database
//
// this is to handle cases where either a third party tool is connected to the database,
// or other Steampipe sessions are attached to an already running Steampipe service
// - we do not want the db service being closed underneath them
//
// note: we need the PgClientAppName check to handle the case where there may be one or more open DB connections
// from this instance at the time of shutdown - for example when a control run is cancelled
// If we do not exclude connections from this execution, the DB will not be shut down after a cancellation
func GetClientCount(ctx context.Context) (*ClientCount, error) {
	putils.LogTime("db_local.GetClientCount start")
	defer putils.LogTime(fmt.Sprintf("db_local.GetClientCount end"))

	rootClient, err := CreateLocalDbConnection(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return nil, err
	}
	defer rootClient.Close(ctx)

	query := `
SELECT 
  application_name,
  count(*)
FROM 
  pg_stat_activity 
WHERE
	-- get only the network client processes
  client_port IS NOT NULL 
	AND
	-- which are client backends
  backend_type=$1 
	AND
	-- which are not connections from this application
  application_name!=$2
GROUP BY application_name
`

	counts := &ClientCount{}

	log.Println("[INFO] ClientConnectionAppName: ", runtime.ClientConnectionAppName)
	rows, err := rootClient.Query(ctx, query, "client backend", runtime.ClientConnectionAppName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var appName string
		var count int

		if err := rows.Scan(&appName, &count); err != nil {
			return nil, err
		}
		log.Printf("[INFO] appName: %s, count: %d", appName, count)

		counts.TotalClients += count

		if db_common.IsClientAppName(appName) {
			counts.SteampipeClients += count
		}

		// plugin manager uses the service prefix
		if db_common.IsServiceAppName(appName) {
			counts.PluginManagerClients += count
		}
	}

	return counts, nil
}

// StopServices searches for and stops the running instance. Does nothing if an instance was not found
func StopServices(ctx context.Context, force bool, invoker constants.Invoker) (status StopStatus, e error) {
	log.Printf("[TRACE] StopDB invoker %s, force %v", invoker, force)
	putils.LogTime("db_local.StopDB start")

	defer func() {
		if e == nil {
			os.Remove(filepaths.RunningInfoFilePath())
		}
		putils.LogTime("db_local.StopDB end")
	}()

	log.Println("[INFO] shutting down plugin manager")
	// stop the plugin manager
	// this means it may be stopped even if we fail to stop the service - that is ok - we will restart it if needed
	pluginManagerStopError := pluginmanager.Stop()
	log.Println("[INFO] shut down plugin manager")

	// stop the DB Service
	log.Println("[INFO] stopping DB Service")
	stopResult, dbStopError := stopDBService(ctx, force)
	log.Println("[INFO] stopped DB Service")

	return stopResult, error_helpers.CombineErrors(dbStopError, pluginManagerStopError)
}

func stopDBService(ctx context.Context, force bool) (StopStatus, error) {
	if force {
		// check if we have a process from another install-dir
		statushooks.SetStatus(ctx, "Checking for running instances…")
		// do not use a context that can be cancelled
		anyStopped := killInstanceIfAny(context.Background())
		if anyStopped {
			return ServiceStopped, nil
		}
		return ServiceNotRunning, nil
	}

	dbState, err := GetState()
	if err != nil {
		return ServiceStopFailed, err
	}

	if dbState == nil {
		// we do not have a info file
		// assume that the service is not running
		return ServiceNotRunning, nil
	}

	// GetStatus has made sure that the process exists
	process, err := psutils.NewProcess(int32(dbState.Pid))
	if err != nil {
		return ServiceStopFailed, err
	}

	err = doThreeStepPostgresExit(ctx, process)
	if err != nil {
		// we couldn't stop it still.
		// timeout
		return ServiceStopTimedOut, err
	}

	return ServiceStopped, nil
}

/*
Postgres has three levels of shutdown:

  - SIGTERM   - Smart Shutdown	 :  Wait for children to end normally - exit self
  - SIGINT    - Fast Shutdown      :  SIGTERM children, causing them to abort current
    transations and exit - wait for children to exit -
    exit self
  - SIGQUIT   - Immediate Shutdown :  SIGQUIT children - wait at most 5 seconds,
    send SIGKILL to children - exit self immediately

Postgres recommended shutdown is to send a SIGTERM - which initiates
a Smart-Shutdown sequence.

IMPORTANT:
As per documentation, it is best not to use SIGKILL
to shut down postgres. Doing so will prevent the server
from releasing shared memory and semaphores.

Reference:
https://www.postgresql.org/docs/12/server-shutdown.html

By the time we actually try to run this sequence, we will have
checked that the service can indeed shutdown gracefully,
the sequence is there only as a backup.
*/
func doThreeStepPostgresExit(ctx context.Context, process *psutils.Process) error {
	putils.LogTime("db_local.doThreeStepPostgresExit start")
	defer putils.LogTime("db_local.doThreeStepPostgresExit end")

	var err error
	var exitSuccessful bool

	// send a SIGTERM
	err = process.SendSignal(syscall.SIGTERM)
	if err != nil {
		return err
	}
	exitSuccessful = waitForProcessExit(process, 2*time.Second)
	if !exitSuccessful {
		// process didn't quit

		// set status, as this is taking time
		statushooks.SetStatus(ctx, "Shutting down…")

		// try a SIGINT
		err = process.SendSignal(syscall.SIGINT)
		if err != nil {
			return err
		}
		exitSuccessful = waitForProcessExit(process, 2*time.Second)
	}
	if !exitSuccessful {
		// process didn't quit
		// desperation prevails
		err = process.SendSignal(syscall.SIGQUIT)
		if err != nil {
			return err
		}
		exitSuccessful = waitForProcessExit(process, 5*time.Second)
	}

	if !exitSuccessful {
		log.Println("[ERROR] Failed to stop service")
		log.Printf("[ERROR] Service Details:\n%s\n", getPrintableProcessDetails(process, 0))
		return fmt.Errorf("service shutdown timed out")
	}

	return nil
}

func waitForProcessExit(process *psutils.Process, waitFor time.Duration) bool {
	putils.LogTime("db_local.waitForProcessExit start")
	defer putils.LogTime("db_local.waitForProcessExit end")

	checkTimer := time.NewTicker(50 * time.Millisecond)
	timeoutAt := time.After(waitFor)

	for {
		select {
		case <-checkTimer.C:
			pEx, _ := utils.PidExists(int(process.Pid))
			if pEx {
				continue
			}
			return true
		case <-timeoutAt:
			checkTimer.Stop()
			return false
		}
	}
}

func getPrintableProcessDetails(process *psutils.Process, indent int) string {
	putils.LogTime("db_local.getPrintableProcessDetails start")
	defer putils.LogTime("db_local.getPrintableProcessDetails end")

	indentString := strings.Repeat("  ", indent)
	appendTo := []string{}

	if name, err := process.Name(); err == nil {
		appendTo = append(appendTo, fmt.Sprintf("%s> Name: %s", indentString, name))
	}
	if cmdLine, err := process.Cmdline(); err == nil {
		appendTo = append(appendTo, fmt.Sprintf("%s> CmdLine: %s", indentString, cmdLine))
	}
	if status, err := process.Status(); err == nil {
		appendTo = append(appendTo, fmt.Sprintf("%s> Status: %s", indentString, status))
	}
	if cwd, err := process.Cwd(); err == nil {
		appendTo = append(appendTo, fmt.Sprintf("%s> CWD: %s", indentString, cwd))
	}
	if executable, err := process.Exe(); err == nil {
		appendTo = append(appendTo, fmt.Sprintf("%s> Executable: %s", indentString, executable))
	}
	if username, err := process.Username(); err == nil {
		appendTo = append(appendTo, fmt.Sprintf("%s> Username: %s", indentString, username))
	}
	if indent == 0 {
		// I do not care about the parent of my parent
		if parent, err := process.Parent(); err == nil && parent != nil {
			appendTo = append(appendTo, "", fmt.Sprintf("%s> Parent Details", indentString))
			parentLog := getPrintableProcessDetails(parent, indent+1)
			appendTo = append(appendTo, parentLog, "")
		}

		// I do not care about all the children of my parent
		if children, err := process.Children(); err == nil && len(children) > 0 {
			appendTo = append(appendTo, fmt.Sprintf("%s> Children Details", indentString))
			for _, child := range children {
				childLog := getPrintableProcessDetails(child, indent+1)
				appendTo = append(appendTo, childLog, "")
			}
		}
	}

	return strings.Join(appendTo, "\n")
}
