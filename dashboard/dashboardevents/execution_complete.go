package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type ExecutionComplete struct {
	Root        dashboardinterfaces.DashboardNodeRun
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionComplete) IsDashboardEvent() {}
