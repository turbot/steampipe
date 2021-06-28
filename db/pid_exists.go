// +build darwin linux

package db

import (
	"fmt"
	"os/exec"
	"strings"
)

// PidExists spawns a subshell with 'ps -p <pid> -o comm='
// and returns true if the process was found - false otherwise
// If there was an error, it'll always return false, whether the process
// exists or not.
func PidExists(pid int) (bool, error) {
	// ps -p 27098 -o comm=
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ps -p %d -o comm=", pid))
	o, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	if strings.Contains(string(o), "(postgres)") {
		// this means that postgres went away, but the process has remained.
		// need to understand why this is happening
		// HACK for now!
		return false, nil
	}
	return (cmd.ProcessState.ExitCode() == 0), nil
}
