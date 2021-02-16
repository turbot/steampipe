package utils

import (
	"fmt"
	"runtime"

	"github.com/containerd/containerd/version"
)

// ConstructUserAgent :: constructs an user-agent string for the installation and env
func ConstructUserAgent(installationID string) string {
	const format = "TURBOT(STEAMPIPE/%s)(%s/%s)(%s/%s)(%s)"

	wsl := "non-wsl"
	isWSL, err := IsWSL()
	if err != nil {
		wsl = "wsl-unknown"
	} else if isWSL {
		wsl = "wsl"
	}

	return fmt.Sprintf(format,
		version.Version,
		runtime.GOOS,
		wsl,
		runtime.GOARCH,
		"",
		installationID)
}
