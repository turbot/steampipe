package dashboardevents

import "github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"

type ExecutionStarted struct {
	Root        dashboardtypes.DashboardNodeRun `json:"dashboard"`
	Panels      map[string]dashboardtypes.SnapshotPanel
	Session     string
	ExecutionId string
	Inputs      map[string]interface{}
	Variables   map[string]string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionStarted) IsDashboardEvent() {}
