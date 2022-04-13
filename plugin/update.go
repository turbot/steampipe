package plugin

import (
	"runtime"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/ociinstaller"
)

func SkipUpdate(report VersionCheckReport) (bool, string) {
	if report.Plugin.ImageDigest == report.CheckResponse.Digest {
		return true, constants.PluginLatestAlreadyInstalled
	}
	if isMacM1() && len(report.Plugin.BinaryArchitecture) == 0 && hasM1Binary(report.CheckResponse.Manifest) {
		return false, ""
	}
	return false, ""
}
func isMacM1() bool {
	return runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
}
func hasM1Binary(manifest responseManifest) bool {
	for _, rml := range manifest.Layers {
		if rml.MediaType == ociinstaller.MediaTypePluginDarwinArm64Layer {
			return true
		}
	}
	return false
}
