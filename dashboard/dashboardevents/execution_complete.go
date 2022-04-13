package dashboardevents

import (
	"time"

	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
)

type ExecutionComplete struct {
	Dashboard   dashboardinterfaces.DashboardNodeRun
	Session     string
	ExecutionId string

	Inputs     map[string]interface{}
	Variables  map[string]string
	SearchPath []string
	StartTime  time.Time
	EndTime    time.Time
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionComplete) IsDashboardEvent() {}
