package dashboardevents

import (
	"github.com/turbot/steampipe/control/controlstatus"
)

type ControlComplete struct {
	ControlName          string
	ControlRunStatus     controlstatus.ControlRunStatus
	ControlStatusSummary *controlstatus.StatusSummary
	Progress             *controlstatus.ControlProgress
	Session              string
	ExecutionId          string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ControlComplete) IsDashboardEvent() {}
