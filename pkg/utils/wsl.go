package utils

import (
	"log"
	"os"
	"runtime"
	"strings"
)

// cache for the WSL value, so that we don't have to query the OS all the time
var isWsl *bool = nil

// IsWSL detects whether app is running in WSL environment
// refer to: https://github.com/Microsoft/WSL/issues/423#issuecomment-679190758
func IsWSL() bool {
	if isWsl != nil {
		return *isWsl
	}
	if runtime.GOOS != "linux" {
		w := false
		isWsl = &w
		return false
	}
	// https://github.com/Microsoft/WSL/issues/2299#issuecomment-361366982
	osReleaseContent, err := os.ReadFile("/proc/version")
	if err != nil {
		log.Println("[TRACE] could not read /proc/version for evaluating WSL: ", err)
		// WSL systems will always have the /proc/version file.
		// if we can't read the file, then this must be some other
		// flavour of linux which doesn't use it - or there's something
		// fundamentally wrong with the installation.
		//
		// in both cases - assume this is not WSL
		return false
	}
	osRelease := strings.ToLower(string(osReleaseContent))
	w := (strings.Contains(osRelease, "microsoft") || strings.Contains(osRelease, "wsl"))
	isWsl = &w
	return *isWsl
}
