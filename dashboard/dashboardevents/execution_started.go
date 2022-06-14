package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardtypes"

type ExecutionStarted struct {
	Root        dashboardtypes.DashboardNodeRun `json:"dashboard"`
	Panels      map[string]dashboardtypes.SnapshotPanel
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionStarted) IsDashboardEvent() {}
