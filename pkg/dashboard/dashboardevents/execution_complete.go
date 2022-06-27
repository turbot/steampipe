package dashboardevents

import (
	"time"

	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
)

type ExecutionComplete struct {
	Root        dashboardtypes.DashboardNodeRun
	Session     string
	ExecutionId string
	Panels      map[string]dashboardtypes.SnapshotPanel
	Inputs      map[string]interface{}
	Variables   map[string]string
	SearchPath  []string
	StartTime   time.Time
	EndTime     time.Time
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionComplete) IsDashboardEvent() {}
