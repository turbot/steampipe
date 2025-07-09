package db_common

import (
	"strings"

	"github.com/turbot/steampipe/v2/pkg/constants"
)

func IsClientAppName(appName string) bool {
	return strings.HasPrefix(appName, constants.ClientConnectionAppNamePrefix) && !strings.HasPrefix(appName, constants.ClientSystemConnectionAppNamePrefix)
}

func IsClientSystemAppName(appName string) bool {
	return strings.HasPrefix(appName, constants.ClientSystemConnectionAppNamePrefix)
}

func IsServiceAppName(appName string) bool {
	return strings.HasPrefix(appName, constants.ServiceConnectionAppNamePrefix)
}
