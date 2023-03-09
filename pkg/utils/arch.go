package utils

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
)

// UnderlyingArch detects the underlying architecture(amd64/arm64) of the system
// we need this to detect the underlying architecture to install the correct FDW package
func UnderlyingArch() (string, error) {
	cmd := exec.Command("uname", "-m")
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	underlyingArch := strings.ToLower(strings.TrimSpace(string(stdout)))

	switch underlyingArch {
	// darwin and linux systems return "x86_64"
	case "x86_64", "amd64":
		return constants.ArchAMD64, nil
	// linux systems return "aarch64"
	case "aarch64", "arm64":
		return constants.ArchARM64, nil
	default:
		return "", fmt.Errorf("Unsupported architecture: %s", underlyingArch)
	}
}

// IsMacM1 returns whether the system is a Mac M1 machine
func IsMacM1() (bool, error) {
	arch, err := UnderlyingArch()
	if err != nil {
		return false, err
	}
	myOs := runtime.GOOS
	isM1 := arch == constants.ArchARM64 && myOs == constants.OSDarwin
	return isM1, nil
}
