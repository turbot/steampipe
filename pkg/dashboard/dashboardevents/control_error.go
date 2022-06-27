package dashboardevents

import "github.com/turbot/steampipe/pkg/control/controlstatus"

type ControlError struct {
	Control     controlstatus.ControlRunStatusProvider
	Progress    *controlstatus.ControlProgress
	Name        string
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ControlError) IsDashboardEvent() {}
