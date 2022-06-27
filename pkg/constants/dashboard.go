package constants

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/version"
)

const (
	DashboardServerDefaultPort    = 9194
	DashboardAssetsImageRefFormat = "us-docker.pkg.dev/steampipe/steampipe/assets:%s"
)

var (
	DashboardAssetsImageRef = fmt.Sprintf(DashboardAssetsImageRefFormat, version.VersionString)
)
