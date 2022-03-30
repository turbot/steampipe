package dashboardevents

import (
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
)

type ExecutionStarted struct {
	Dashboard   dashboardinterfaces.DashboardNodeRun `json:"dashboard"`
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionStarted) IsDashboardEvent() {}
