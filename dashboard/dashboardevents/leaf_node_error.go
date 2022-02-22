package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type LeafNodeError struct {
	LeafNode dashboardinterfaces.DashboardNodeRun
	Session  string
	Error    error
}

// IsDashboardEvent implements DashboardEvent interface
func (*LeafNodeError) IsDashboardEvent() {}
