package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// UnderlyingArch detects the underlying architecture(amd64/arm64) of the system
func UnderlyingArch() string {
	name := "uname"
	arg0 := "-m"

	cmd := exec.Command(name, arg0)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Errorf("Error finding the underlying architecture: %s", err.Error())
	}
	underlyingArch := strings.ToLower(strings.TrimSpace(string(stdout)))

	switch underlyingArch {
	// darwin and linux systems return "x86_64"
	case "x86_64":
		return "amd64"
	// linux systems return "aarch64"
	case "aarch64":
		return "arm64"
	default:
		return underlyingArch
	}
}
