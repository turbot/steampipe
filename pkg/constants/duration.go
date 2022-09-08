package constants

import "time"

var (
	DashboardServiceStartTimeout = 30 * time.Second
	DBConnectionTimeout          = 5 * time.Second
	ServicePingInterval          = 50 * time.Millisecond
)
