package utils

import (
	"io/ioutil"
	"runtime"
	"strings"
)

// IsWSL :: detects whether app is running in WSL environment
// refer to: https://github.com/Microsoft/WSL/issues/423#issuecomment-679190758
func IsWSL() (bool, error) {
	if runtime.GOOS != "linux" {
		return false, nil
	}
	osReleaseContent, err := ioutil.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false, err
	}
	osRelease := strings.ToLower(string(osReleaseContent))
	return (strings.Contains(osRelease, "microsoft") || strings.Contains(osRelease, "wsl")), nil
}
