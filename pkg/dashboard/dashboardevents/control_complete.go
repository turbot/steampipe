package dashboardevents

import (
	"github.com/turbot/steampipe/pkg/control/controlstatus"
)

type ControlComplete struct {
	Progress    *controlstatus.ControlProgress
	Control     controlstatus.ControlRunStatusProvider
	Name        string
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ControlComplete) IsDashboardEvent() {}
