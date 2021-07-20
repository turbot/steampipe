package db

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/steampipe/display"

	"github.com/turbot/steampipe/utils"
)

// StopStatus :: pseudoEnum for service stop result
type StopStatus int

const (
	// ServiceStopped :: StopStatus - service was stopped
	ServiceStopped StopStatus = iota
	// ServiceNotRunning :: StopStatus - service was not running
	ServiceNotRunning
	// ServiceStopFailed :: StopStatus - service could not be stopped
	ServiceStopFailed
	// ServiceStopTimedOut :: StopStatus - service stop attempt timed out
	ServiceStopTimedOut
)

// Shutdown :: closes the client connection and stops the
// database instance if the given `invoker` matches
func Shutdown(client *Client, invoker Invoker) {
	utils.LogTime("db.Shutdown start")
	defer utils.LogTime("db.Shutdown end")

	if client != nil {
		client.Close()
	}

	status, _ := GetStatus()

	// is the service running?
	if status != nil {
		if status.Invoker == InvokerService {
			// if the service was invoked by `steampipe service`,
			// then we don't shut it down
			return
		}

		count, _ := GetCountOfConnectedClients()

		if count > 0 {
			// there are other clients connected to the database
			// we can't stop the DB.
			return
		}

		// we can shutdown the database
		status, err := StopDB(false, invoker, nil)
		if err != nil {
			utils.ShowError(err)
		}
		if status != ServiceStopped {
			StopDB(true, invoker, nil)
		}
	}
}

func GetCountOfConnectedClients() (int, error) {
	rootClient, err := createRootDbClient()
	if err != nil {
		return -1, err
	}
	row := rootClient.QueryRow("select count(*) from pg_stat_activity where client_port IS NOT NULL and application_name='steampipe' and backend_type='client backend';")
	count := 0
	row.Scan(&count)
	rootClient.Close()
	return (count - 1 /* deduct the existing client */), nil
}

// StopDB :: search and stop the running instance. Does nothing if an instance was not found
func StopDB(force bool, invoker Invoker, spinner *spinner.Spinner) (StopStatus, error) {
	log.Println("[TRACE] StopDB", force)

	utils.LogTime("db.StopDB start")
	defer utils.LogTime("db.StopDB end")

	if force {
		// remove this file regardless of whether
		// we could stop the service or not
		// so that the next time it starts,
		// all previous instances are nuked
		defer os.Remove(runningInfoFilePath())
	}
	info, err := GetStatus()
	if err != nil {
		return ServiceStopFailed, err
	}

	if force {
		// check if we have a process from another install-dir
		display.UpdateSpinnerMessage(spinner, "Checking for running instances...")
		killInstanceIfAny()
		return ServiceStopped, nil
	}

	if info == nil {
		// we do not have a info file
		// assume that the service is not running
		return ServiceNotRunning, nil
	}

	// GetStatus has made sure that the process exists
	process, err := psutils.NewProcess(int32(info.Pid))
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
	Postgres has two more levels of shutdown:
		* SIGTERM	- Smart Shutdown    	:  Wait for children to end normally - exit self
		* SIGINT	- Fast Shutdown      	:  SIGTERM children, causing them to abort current
											:  transations and exit - wait for children to exit -
											:  exit self
		* SIGQUIT	- Immediate Shutdown 	:  SIGQUIT children - wait at most 5 seconds,
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
		log.Printf("[ERROR] ** > Failed to stop service\n")
		errorLogProcessDetails(process, 0)
		log.Printf("[ERROR] ** > Done\n")
		return fmt.Errorf("service shutdown timed out")
	}

	return nil
}

func waitForProcessExit(process *psutils.Process, waitFor time.Duration) bool {
	checkTimer := time.NewTicker(50 * time.Millisecond)
	timeoutAt := time.After(waitFor)

	for {
		select {
		case <-checkTimer.C:
			pEx, _ := PidExists(int(process.Pid))
			if pEx {
				continue
			}
			return true
		case <-timeoutAt:
			return false
		}
	}
}

func errorLogProcessDetails(process *psutils.Process, indent int) {
	indentString := strings.Repeat("  ", indent)

	if name, err := process.Name(); err == nil {
		log.Printf("[ERROR] ** %s > Name: %s\n", indentString, name)
	}
	if cmdLine, err := process.Cmdline(); err == nil {
		log.Printf("[ERROR] ** %s > CmdLine: %s\n", indentString, cmdLine)
	}
	if status, err := process.Status(); err == nil {
		log.Printf("[ERROR] ** %s > Status: %s\n", indentString, status)
	}
	if cwd, err := process.Cwd(); err == nil {
		log.Printf("[ERROR] ** %s > CWD: %s\n", indentString, cwd)
	}
	if executable, err := process.Exe(); err == nil {
		log.Printf("[ERROR] ** %s > Executable: %s\n", indentString, executable)
	}
	if username, err := process.Username(); err == nil {
		log.Printf("[ERROR] ** %s > Username: %s\n", indentString, username)
	}
	if indent == 0 {
		// I do not care about the parent of my parent
		if parent, err := process.Parent(); err == nil && parent != nil {
			log.Printf("[ERROR] ** %s > Parent Details\n", indentString)
			errorLogProcessDetails(parent, indent+1)
		}

		// I do not care about all the children of my parent
		if children, err := process.Children(); err == nil && len(children) > 0 {
			log.Printf("[ERROR] ** %s > Children Details\n", indentString)
			for _, child := range children {
				errorLogProcessDetails(child, indent+1)
			}
		}
	}
	log.Printf("[ERROR] ** \n")
}
