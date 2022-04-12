package dashboardevents

import (
	"time"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
)

type ExecutionComplete struct {
	Dashboard   dashboardinterfaces.DashboardNodeRun
	Session     string
	ExecutionId string

	Inputs     map[string]interface{}
	Variables  map[string]*modconfig.Variable
	SearchPath []string
	StartTime  time.Time
	EndTime    time.Time
	Actor      string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionComplete) IsDashboardEvent() {}
