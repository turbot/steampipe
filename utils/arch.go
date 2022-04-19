package utils

import (
	"fmt"
	"os/exec"
	"strings"
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
		return "amd64", nil
	// linux systems return "aarch64"
	case "aarch64", "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("Unsupported architecture: %s", underlyingArch)
	}
}
