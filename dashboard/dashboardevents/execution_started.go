package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type ExecutionStarted struct {
	DashboardNode dashboardinterfaces.DashboardNodeRun `json:"dashboard"`
	Session       string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionStarted) IsDashboardEvent() {}
