package dashboardevents

import "github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"

type LeafNodeComplete struct {
	LeafNode    dashboardtypes.DashboardTreeRun
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*LeafNodeComplete) IsDashboardEvent() {}
