package db

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/turbot/steampipe/constants"

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

// StopDB :: search and stop the running instance. Does nothing if an instance was not found
func StopDB(force bool) (StopStatus, error) {
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

	if info == nil {
		// we do not have a info file
		return ServiceNotRunning, nil
	}

	doesPidExist, err := pidExists(info.Pid)

	if err != nil {
		return ServiceStopFailed, err
	}

	if !doesPidExist {
		// nothing to do here
		return ServiceNotRunning, os.Remove(runningInfoFilePath())
	}

	if info.Invoker != InvokerService && !force {
		return ServiceStopFailed, fmt.Errorf("You have a %s session open. Close this session before running %s.\nTo kill existing sessions, run %s", constants.Bold("steampipe query"), constants.Bold("steampipe service stop"), constants.Bold("steampipe service stop --force"))
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
			pEx, err := pidExists(info.Pid)
			if err != nil {
				utils.ShowError(err)
			}
			if err == nil && !pEx {
				// no more process
				processKilledChannel <- "killed"
				break
			}
			if time.Since(signalSentAt) > constants.SpinnerShowTimeout && !spinnerShown {
				s := utils.ShowSpinner("Shutting down...")
				defer utils.StopSpinner(s)
				spinnerShown = true
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	timeoutAfter := time.After(10 * time.Second)

	select {
	case <-timeoutAfter:
		log.Println("[TRACE] timed out", force)
		if force {
			// timed out, we couldn't manage to stop it!
			// do a SIGQUIT
			log.Println("[TRACE] force killing")
			err = process.Kill()
			if err != nil {
				log.Println("[TRACE] could not force kill", err)
				utils.ShowError(err)
			}
			os.Remove(runningInfoFilePath())
			return ServiceStopped, nil
		}
		return ServiceStopTimedOut, nil
	case <-processKilledChannel:
		os.Remove(runningInfoFilePath())
		return ServiceStopped, nil
	}
}
