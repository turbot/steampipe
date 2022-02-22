package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type LeafNodeComplete struct {
	Node    dashboardinterfaces.DashboardNodeRun
	Session string
}

// IsDashboardEvent implements DashboardEvent interface
func (*LeafNodeComplete) IsDashboardEvent() {}
