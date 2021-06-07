package db

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
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
	log.Println("[TRACE] shutdown")
	if client != nil {
		client.Close()
	}

	status, _ := GetStatus()

	// force stop if the service was invoked by the same invoker and we are the last one
	if status != nil && status.Invoker == invoker {
		status, err := StopDB(false, invoker)
		if err != nil {
			utils.ShowError(err)
		}
		if status != ServiceStopped {
			StopDB(true, invoker)
		}
	}
}

// StopDB :: search and stop the running instance. Does nothing if an instance was not found
func StopDB(force bool, invoker Invoker) (StopStatus, error) {
	log.Println("[TRACE] StopDB", force)

	if force {
		// remove this file regardless of whether
		// we could stop the service or not
		// so that the next time it starts,
		// all previous instances are nuked
		defer os.Remove(runningInfoFilePath())
	}

	info, err := loadRunningInstanceInfo()
	if err != nil {
		return ServiceStopFailed, err
	}

	if force {
		// check if we have a process from another install-dir
		checkedPreviousInstances := make(chan bool, 1)
		s := display.StartSpinnerAfterDelay("Checking for running instances...", constants.SpinnerShowTimeout, checkedPreviousInstances)
		for {
			previousProcess := findSteampipePostgresInstance()
			if previousProcess != nil {
				// we have an errant process
				killProcessTree(previousProcess)
				continue
			}
			break
		}
		close(checkedPreviousInstances)
		display.StopSpinner(s)
		os.Remove(runningInfoFilePath())
		return ServiceStopped, nil
	}

	if info == nil {
		// we do not have a info file
		return ServiceNotRunning, nil
	}

	doesPidExist, err := psutils.PidExists(int32(info.Pid))

	if err != nil {
		return ServiceStopFailed, err
	}

	if !doesPidExist {
		// nothing to do here
		return ServiceNotRunning, os.Remove(runningInfoFilePath())
	}

	if info.Invoker != invoker {
		return ServiceStopFailed, fmt.Errorf("You have a %s session open. The service will be stopped when the session ends.\nTo kill existing sessions, run %s", constants.Bold(fmt.Sprintf("steampipe %s", info.Invoker)), constants.Bold("steampipe service stop --force"))
	}

	process, err := os.FindProcess(info.Pid)
	if err != nil {
		return ServiceStopFailed, err
	}

	// we need to do this since the standard
	// cmd.Process.Kill() sends a SIGKILL which
	// makes PG terminate immediately without saving state.
	// refer: https://www.postgresql.org/docs/12/server-shutdown.html
	// refer: https://golang.org/src/os/exec_posix.go?h=kill#L65
	killSignal := syscall.SIGTERM
	if force {
		killSignal = syscall.SIGINT
	}

	err = process.Signal(killSignal)
	log.Println("[TRACE] signal sent", killSignal)

	if err != nil {
		return ServiceStopFailed, err
	}

	signalSentAt := time.Now()
	spinnerShown := false

	processKilledChannel := make(chan string, 1)
	go func() {
		for {
			pEx, err := psutils.PidExists(int32(info.Pid))
			if err != nil {
				utils.ShowError(err)
			}
			if err == nil && !pEx {
				// no more process
				processKilledChannel <- "killed"
				break
			}
			if time.Since(signalSentAt) > constants.SpinnerShowTimeout && !spinnerShown {
				if cmdconfig.Viper().GetBool(constants.ConfigKeyShowInteractiveOutput) {
					s := display.ShowSpinner("Shutting down...")
					defer display.StopSpinner(s)
					spinnerShown = true
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	timeoutAfter := time.After(10 * time.Second)

	select {
	case <-timeoutAfter:
		return ServiceStopTimedOut, nil
	case <-processKilledChannel:
		os.Remove(runningInfoFilePath())
		return ServiceStopped, nil
	}
}
