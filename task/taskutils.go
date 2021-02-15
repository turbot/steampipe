package task

import (
	"fmt"
	"runtime"
)

func constructUserAgent(installationID string) string {
	const format = "TURBOT(STEAMPIPE/%s)(%s/%s)(%s/%s)(%s)"

	return fmt.Sprintf(format,
		currentVersion,
		runtime.GOOS,
		"",
		runtime.GOARCH,
		"",
		installationID)
}

const disableUpdatesCheckEnvVar = "SP_DISABLE_UPDATE_CHECK"
