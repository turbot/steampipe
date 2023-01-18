package constants

import "time"

var (
	DashboardServiceStartTimeout = 30 * time.Second
	DBConnectionTimeout          = 30 * time.Second
	DBConnectionRetryBackoff     = 200 * time.Millisecond
	ServicePingInterval          = 50 * time.Millisecond
)
