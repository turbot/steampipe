package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type LeafNodeError struct {
	LeafNode    dashboardinterfaces.DashboardNodeRun
	Session     string
	Error       error
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*LeafNodeError) IsDashboardEvent() {}
