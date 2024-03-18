package plugin

import (
	"runtime"

	"github.com/turbot/steampipe/pkg/constants"
)

// UpdateRequired determines if the latest version in a "stream"
// requires the plugin to update.
func UpdateRequired(report VersionCheckReport) bool {

	// 1) If there is an updated version ALWAYS update
	if report.Plugin.Version != report.CheckResponse.Version {
		return true
	}

	// 2) If we are M1, current installed version is AMD, and ARM is available - update
	if isRunningAsMacM1() && report.Plugin.BinaryArchitecture != constants.ArchARM64 {
		return true
	}

	// 3) Otherwise skip
	return false
}

// check to see if steampipe is running as a Mac/M1 build
// Mac/M1 can run 'amd64' builds, but that is not a
// problem, since they will be running under 'rosetta'
// TODO: Find a way to determine the underlying architecture, rather than depending on Go runtime
func isRunningAsMacM1() bool {
	return runtime.GOOS == constants.OSDarwin && runtime.GOARCH == constants.ArchARM64
}
