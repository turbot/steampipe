package db_local

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/plugin_manager"

	"github.com/turbot/steampipe/utils"
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
func ShutdownService(invoker constants.Invoker) {
	utils.LogTime("db_local.ShutdownService start")
	defer utils.LogTime("db_local.ShutdownService end")

	status, _ := GetState()

	// if the service is not running or it was invoked by 'steampipe service',
	// then we don't shut it down
	if status == nil || status.Invoker == constants.InvokerService {
		return
	}

	// how many clients are connected
	// under a fresh context
	count, _ := GetCountOfConnectedClients(context.Background())
	if count > 0 {
		// there are other clients connected to the database
		// we can't stop the DB.
		return
	}

	// we can shut down the database
	stopStatus, err := StopServices(false, invoker, nil)
	if err != nil {
		utils.ShowError(err)
	}
	if stopStatus == ServiceStopped {
		return
	}

	// shutdown failed - try to force stop
	_, err = StopServices(true, invoker, nil)
	if err != nil {
		utils.ShowError(err)
	}

}

// GetCountOfConnectedClients returns the number of clients currently connected to the service
func GetCountOfConnectedClients(ctx context.Context) (i int, e error) {
	utils.LogTime("db_local.GetCountOfConnectedClients start")
	defer utils.LogTime(fmt.Sprintf("db_local.GetCountOfConnectedClients end:%d", i))

	rootClient, err := createLocalDbClient(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return -1, err
	}
	defer rootClient.Close()

	clientCount := 0
	// get the total number of connected clients
	row := rootClient.QueryRow("select count(*) from pg_stat_activity where client_port IS NOT NULL and backend_type='client backend';")
	row.Scan(&clientCount)
	// clientCount can never be zero, since the client we are using to run the query counts as a client
	// deduct the open connections in the pool of this client
	return clientCount - rootClient.Stats().OpenConnections, nil
}

// StopServices searches for and stops the running instance. Does nothing if an instance was not found
func StopServices(force bool, invoker constants.Invoker, spinner *spinner.Spinner) (status StopStatus, e error) {
	log.Printf("[TRACE] StopDB invoker %s, force %v", invoker, force)
	utils.LogTime("db_local.StopDB start")

	defer func() {
		if e == nil {
			os.Remove(constants.RunningInfoFilePath())
		}
		utils.LogTime("db_local.StopDB end")
	}()

	// stop the plugin manager
	// this means it may be stopped even if we fail to stop the service - that is ok - we will restart it if needed
	pluginManagerStopError := plugin_manager.Stop()

	// stop the DB Service
	stopResult, dbStopError := stopDBService(spinner, force)

	return stopResult, utils.CombineErrors(dbStopError, pluginManagerStopError)
}

func stopDBService(spinner *spinner.Spinner, force bool) (StopStatus, error) {
	if force {
		// check if we have a process from another install-dir
		display.UpdateSpinnerMessage(spinner, "Checking for running instances...")
		// do not use a context that can be cancelled
		killInstanceIfAny(context.Background())
		return ServiceStopped, nil
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

	display.UpdateSpinnerMessage(spinner, "Shutting down...")

	err = doThreeStepPostgresExit(process)
	if err != nil {
		// we couldn't stop it still.
		// timeout
		return ServiceStopTimedOut, err
	}

	return ServiceStopped, nil
}

/**
	Postgres has three levels of shutdown:

	* SIGTERM   - Smart Shutdown	 :  Wait for children to end normally - exit self
	* SIGINT    - Fast Shutdown      :  SIGTERM children, causing them to abort current
									    transations and exit - wait for children to exit -
									    exit self
	* SIGQUIT   - Immediate Shutdown :  SIGQUIT children - wait at most 5 seconds,
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
**/
func doThreeStepPostgresExit(process *psutils.Process) error {
	utils.LogTime("db_local.doThreeStepPostgresExit start")
	defer utils.LogTime("db_local.doThreeStepPostgresExit end")

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
	utils.LogTime("db_local.waitForProcessExit start")
	defer utils.LogTime("db_local.waitForProcessExit end")

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
	utils.LogTime("db_local.getPrintableProcessDetails start")
	defer utils.LogTime("db_local.getPrintableProcessDetails end")

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
