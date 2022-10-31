package utils

import (
	"os"
	"runtime"
	"strings"
)

// cache for the WSL value, so that we don't have to query the OS all the time
var wslValue *bool = nil

// IsWSL :: detects whether app is running in WSL environment
// refer to: https://github.com/Microsoft/WSL/issues/423#issuecomment-679190758
func IsWSL() (bool, error) {
	if wslValue != nil {
		return *wslValue, nil
	}
	if runtime.GOOS != "linux" {
		return false, nil
	}
	// https://github.com/Microsoft/WSL/issues/2299#issuecomment-361366982
	osReleaseContent, err := os.ReadFile("/proc/version")
	if err != nil {
		return false, err
	}
	osRelease := strings.ToLower(string(osReleaseContent))
	w := (strings.Contains(osRelease, "microsoft") || strings.Contains(osRelease, "wsl"))
	wslValue = &w
	return *wslValue, nil
}
