package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type ExecutionComplete struct {
	Dashboard dashboardinterfaces.DashboardNodeRun
	Session   string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionComplete) IsDashboardEvent() {}
