package dashboardevents

import (
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
)

type ExecutionStarted struct {
	Root        dashboardinterfaces.DashboardNodeRun `json:"dashboard"`
	LeafNodes   map[string]dashboardinterfaces.SnapshotPanel
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionStarted) IsDashboardEvent() {}
