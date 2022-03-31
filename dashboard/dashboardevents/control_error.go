package dashboardevents

import "github.com/turbot/steampipe/control/controlstatus"

type ControlError struct {
	ControlName          string
	ControlRunStatus     controlstatus.ControlRunStatus
	ControlStatusSummary *controlstatus.StatusSummary
	Progress             *controlstatus.ControlProgress
	Session              string
	ExecutionId          string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ControlError) IsDashboardEvent() {}
