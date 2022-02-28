package constants

import (
	"fmt"

	"github.com/turbot/steampipe/version"
)

const (
	DashboardServerDefaultPort    = 9194
	DashboardAssetsImageRefFormat = "us-docker.pkg.dev/steampipe/steampipe/assets:%s"
)

func DashboardAssetsImageRef() string {
	return fmt.Sprintf(DashboardAssetsImageRefFormat, version.VersionString)
}
