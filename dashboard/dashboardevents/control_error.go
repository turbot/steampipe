package dashboardevents

import "github.com/turbot/steampipe/control/controlstatus"

type ControlError struct {
	Control     controlstatus.ControlRunStatusProvider
	Progress    *controlstatus.ControlProgress
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ControlError) IsDashboardEvent() {}
