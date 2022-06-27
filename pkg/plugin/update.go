package plugin

import (
	"runtime"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/ociinstaller"
)

// SkipUpdate determines if the latest version in a "stream"
// requires the plugin to update.
func SkipUpdate(report VersionCheckReport) (bool, string) {

	// 1) If there is an updated version ALWAYS update
	if report.Plugin.ImageDigest != report.CheckResponse.Digest {
		return false, ""
	}

	// 2) If we are M1, current installed version is AMD, and ARM is available - update
	if isRunningAsMacM1() && manifestHasM1Binary(report.CheckResponse.Manifest) && report.Plugin.BinaryArchitecture != constants.ArchARM64 {
		return false, ""
	}

	// 3) Otherwise skip
	return true, constants.PluginLatestAlreadyInstalled
}

// check to see if steampipe is running as a Mac/M1 build
// Mac/M1 can run 'amd64' builds, but that is not a
// problem, since they will be running under 'rosetta'
// TODO: Find a way to determine the underlying architecture, rather than depending on Go runtime
func isRunningAsMacM1() bool {
	return runtime.GOOS == constants.OSDarwin && runtime.GOARCH == constants.ArchARM64
}

func manifestHasM1Binary(manifest responseManifest) bool {
	for _, rml := range manifest.Layers {
		if rml.MediaType == ociinstaller.MediaTypePluginDarwinArm64Layer {
			return true
		}
	}
	return false
}
