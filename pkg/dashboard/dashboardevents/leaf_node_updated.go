package dashboardevents

import "github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"

type LeafNodeUpdated struct {
	LeafNode    dashboardtypes.DashboardTreeRun
	Session     string
	ExecutionId string
}

func NewLeafNodeUpdate(r dashboardtypes.DashboardTreeRun, session, executionId string) *LeafNodeUpdated {
	return &LeafNodeUpdated{
		LeafNode:    r,
		Session:     session,
		ExecutionId: executionId,
	}
}

// IsDashboardEvent implements DashboardEvent interface
func (*LeafNodeUpdated) IsDashboardEvent() {}
