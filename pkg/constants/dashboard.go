package constants

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/viper"
)

// DashboardListenAddresses is an arrays is listen addresses which Steampipe accepts
var DashboardListenAddresses = []string{"localhost", "127.0.0.1"}

const (
	DashboardServerDefaultPort    = 9194
	DashboardAssetsImageRefFormat = "us-docker.pkg.dev/steampipe/steampipe/assets:%s"
)

var (
	DashboardAssetsImageRef = fmt.Sprintf(DashboardAssetsImageRefFormat, semver.MustParse(viper.GetString("main.version")))
)
