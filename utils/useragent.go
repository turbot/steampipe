package utils

import (
	"fmt"
	"runtime"

	"github.com/containerd/containerd/version"
)

func ConstructUserAgent(installationID string) string {

	wslString := ""
	if runtime.GOOS != "linux" {
		wslString = "wsl-na"
	} else {
		wsl, err := IsWSL()
		if err != nil {
			wslString = "wsl-unknown"
		} else if wsl {
			wslString = "wsl-win"
		} else {
			wslString = "wsl-nil"
		}

	}

	// TURBOT(STEAMPIPE/1.4.1+unknown)(linux/wsl-nil)(amd64)(wsl-test)
	const format = "TURBOT(STEAMPIPE/%s)(%s/%s)(%s)(%s)"

	return fmt.Sprintf(format,
		version.Version,
		runtime.GOOS,
		wslString,
		runtime.GOARCH,
		installationID)
}
