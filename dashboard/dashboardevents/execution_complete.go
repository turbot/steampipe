package dashboardevents

import (
	"time"

	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
)

type ExecutionComplete struct {
	Root        dashboardinterfaces.DashboardNodeRun
	Session     string
	ExecutionId string
	Panels      map[string]dashboardinterfaces.SnapshotPanel
	Inputs      map[string]interface{}
	Variables   map[string]string
	SearchPath  []string
	StartTime   time.Time
	EndTime     time.Time
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionComplete) IsDashboardEvent() {}
