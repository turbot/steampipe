package utils

import (
	"fmt"
	"runtime"

	"github.com/containerd/containerd/version"
)

func ConstructUserAgent(installationID string) string {
	const format = "TURBOT(STEAMPIPE/%s)(%s/%s)(%s/%s)(%s)"

	return fmt.Sprintf(format,
		version.Version,
		runtime.GOOS,
		"",
		runtime.GOARCH,
		"",
		installationID)
}
