package dashboardevents

import (
	"github.com/turbot/steampipe/control/controlstatus"
)

type ControlComplete struct {
	Progress    *controlstatus.ControlProgress
	Control     controlstatus.ControlRunStatusProvider
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ControlComplete) IsDashboardEvent() {}
