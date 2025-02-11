package utils

import (
	"fmt"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/pipe-fittings/v2/utils"
)

// TODO We should look to use pipe-fittings/v2/utils.PidExists instead of this function.
// Currently when using the pipe-fittings function, we are seeing some errors with the pids not being found
// resulting in unsuccessful service shutdowns.
// https://github.com/turbot/steampipe/issues/4487

// PidExists scans through the list of PIDs in the system
// and checks for the `targetPID`.
//
// PidExists uses iteration, instead of signalling, since we have observed that
// signalling does not always work reliably when the destination of the signal
// is a child of the source of the signal - which may be the case then starting
// implicit services
func PidExists(targetPid int) (bool, error) {
	utils.LogTime("utils.PidExists start")
	defer utils.LogTime("utils.PidExists end")

	process, err := FindProcess(targetPid)
	found := process != nil
	return found, err
}

// FindProcess tries to find the process with the given pid
// returns nil if the process could not be found
func FindProcess(targetPid int) (*psutils.Process, error) {
	utils.LogTime("utils.FindProcess start")
	defer utils.LogTime("utils.FindProcess end")

	pids, err := psutils.Pids()
	if err != nil {
		return nil, fmt.Errorf("failed to get pids")
	}
	for _, pid := range pids {
		if targetPid == int(pid) {
			//nolint: gosec	// target pdi will be 32 bit
			process, err := psutils.NewProcess(int32(targetPid))
			if err != nil {
				return nil, nil
			}

			status, err := process.Status()
			if err != nil {
				return nil, fmt.Errorf("failed to get status: %s", err.Error())
			}

			if status == "Z" {
				// this means that postgres went away, but the process itself is still a zombie.
				return nil, nil
			}
			return process, nil
		}
	}
	return nil, nil
}
