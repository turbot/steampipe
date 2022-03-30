package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type LeafNodeProgress struct {
	LeafNode    dashboardinterfaces.DashboardNodeRun
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*LeafNodeProgress) IsDashboardEvent() {}
