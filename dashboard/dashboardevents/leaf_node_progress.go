package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type LeafNodeProgress struct {
	Node    dashboardinterfaces.DashboardNodeRun
	Session string
}

// IsDashboardEvent implements DashboardEvent interface
func (*LeafNodeProgress) IsDashboardEvent() {}
