// +build darwin linux

package db

import (
	"strings"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/steampipe/utils"
)

// PidExists scans through the list of PIDs in the system
// and checks for the `targetPID`.
//
// PidExists uses iteration, instead of signalling, since we have observed that
// signalling does not always work reliably when the destination of the signal
// is a child of the source of the signal - which may be the case then starting
// implicit services
//
func PidExists(targetPid int) (bool, error) {
	utils.LogTime("db.PidExists start")
	defer utils.LogTime("db.PidExists end")

	pids, err := psutils.Pids()
	if err != nil {
		return false, nil
	}
	for _, pid := range pids {
		if targetPid == int(pid) {
			process, err := psutils.NewProcess(int32(targetPid))
			if err != nil {
				return true, nil
			}
			cmdLine, err := process.Cmdline()
			if err != nil {
				return true, err
			}
			if strings.Contains(cmdLine, "(postgres)") {
				// this means that postgres went away, but the process has not yet completed.
				// we are not sure why this occurs but can safely treat it as if the process does not exist
				return false, nil
			}
			return true, nil
		}
	}
	return false, nil
}
