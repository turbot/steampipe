package db_local

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"syscall"
	"time"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
)

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
	utils.LogTime("db_local.doThreeStepPostgresExit start")
	defer utils.LogTime("db_local.doThreeStepPostgresExit end")

	var err error
	var exitSuccessful bool

	// send a SIGTERM
	err = process.SendSignal(syscall.SIGTERM)
	if err != nil {
		return err
	}
	exitSuccessful = waitForProcessExit(ctx, process)
	if !exitSuccessful {
		// process didn't quit

		// set status, as this is taking time
		statushooks.SetStatus(ctx, "Shutting down…")

		// try a SIGINT
		err = process.SendSignal(syscall.SIGINT)
		if err != nil {
			return err
		}
		exitSuccessful = waitForProcessExit(ctx, process)
	}
	if !exitSuccessful {
		// process didn't quit
		// desperation prevails
		err = process.SendSignal(syscall.SIGQUIT)
		if err != nil {
			return err
		}
		exitSuccessful = waitForProcessExit(ctx, process)
	}

	if !exitSuccessful {
		log.Println("[ERROR] Failed to stop service")
		log.Printf("[ERROR] Service Details:\n%s\n", getPrintableProcessDetails(process, 0))
		return fmt.Errorf("service shutdown timed out")
	}

	return nil
}

func waitForProcessExit(ctx context.Context, process *psutils.Process) bool {
	utils.LogTime("db_local.waitForProcessExit start")
	defer utils.LogTime("db_local.waitForProcessExit end")

	checkTimer := time.NewTicker(50 * time.Millisecond)

	for {
		select {
		case <-checkTimer.C:
			pEx, _ := utils.PidExists(int(process.Pid))
			if pEx {
				continue
			}
			return true
		case <-ctx.Done():
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

// kill all postgres processes that were started as part of steampipe (if any)
func killPostgresInstanceIfAny(ctx context.Context) bool {
	processes, err := FindAllSteampipePostgresInstances(ctx)
	if err != nil {
		return false
	}
	wg := sync.WaitGroup{}
	for _, process := range processes {
		wg.Add(1)
		go func(p *psutils.Process) {
			doThreeStepPostgresExit(ctx, p)
			wg.Done()
		}(process)
	}
	wg.Wait()
	return len(processes) > 0
}

// kill all plugin manager processes that were started as part of steampipe (if any)
func killPluginManagerInstanceIfAny(ctx context.Context) bool {
	processGroups, err := FindAllSteampipePluginManagerInstances(ctx)
	if err != nil {
		return false
	}
	wg := sync.WaitGroup{}
	for _, processGroup := range processGroups {
		// add the number of processes in this to the waitGroup
		wg.Add(len(processGroup))
		go func(processGrp []*psutils.Process) {
			for _, p := range processGrp {
				defer wg.Done()
				if err := p.KillWithContext(ctx); err != nil {
					log.Println("[TRACE] error killing process", err)
				}
				timeout, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()
				if !waitForProcessExit(timeout, p) {
					log.Println("[TRACE] timed out waiting for process exit")
				}
			}
		}(processGroup)
	}
	wg.Wait()
	return len(processGroups) > 0
}

func FindAllSteampipePostgresInstances(ctx context.Context) ([]*psutils.Process, error) {
	var instances []*psutils.Process
	allProcesses, err := psutils.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range allProcesses {
		cmdLine, err := p.CmdlineSliceWithContext(ctx)
		if err != nil {
			return nil, err
		}
		if isSteampipePostgresProcess(ctx, cmdLine) {
			instances = append(instances, p)
		}
	}
	return instances, nil
}

func FindAllSteampipePluginManagerInstances(ctx context.Context) ([][]*psutils.Process, error) {
	var instances [][]*psutils.Process
	allProcesses, err := psutils.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range allProcesses {
		cmdLine, err := p.CmdlineSliceWithContext(ctx)
		if err != nil {
			return nil, err
		}
		if isSteampipePluginManagerProcess(ctx, cmdLine) {
			theseInstances := []*psutils.Process{}
			for _, q := range allProcesses {
				if ppid, err := q.Ppid(); err == nil && ppid == p.Pid {
					// add all child plugin processes too
					theseInstances = append(theseInstances, q)
				}
			}
			theseInstances = append(theseInstances, p)
			instances = append(instances, theseInstances)
		}
	}
	return instances, nil
}

func isSteampipePostgresProcess(ctx context.Context, cmdline []string) bool {
	if len(cmdline) < 1 {
		return false
	}
	if strings.Contains(cmdline[0], "postgres") {
		// this is a postgres process - but is it a steampipe service?
		return helpers.StringSliceContains(cmdline, fmt.Sprintf("application_name=%s", constants.AppName))
	}
	return false
}

func isSteampipePluginManagerProcess(ctx context.Context, cmdline []string) bool {
	if len(cmdline) < 1 {
		return false
	}
	return strings.HasSuffix(cmdline[0], "steampipe") && strings.EqualFold(cmdline[1], "plugin-manager")
}
