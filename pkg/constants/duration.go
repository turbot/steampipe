package constants

import "time"

var (
	DashboardStartTimeout    = 30 * time.Second
	DBStartTimeout           = 30 * time.Second
	DBConnectionRetryBackoff = 200 * time.Millisecond
	DBRecoveryTimeout        = 24 * time.Hour
	DBRecoveryRetryBackoff   = 200 * time.Millisecond
	ServicePingInterval      = 50 * time.Millisecond
	PluginStartTimeout       = 30 * time.Second
)
