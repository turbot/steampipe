package dashboardevents

import (
	"time"

	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/inputvars"
)

type ExecutionComplete struct {
	Dashboard   dashboardinterfaces.DashboardNodeRun
	Session     string
	ExecutionId string

	Inputs     map[string]interface{}
	Variables  map[string]*inputvars.InputValue
	SearchPath []string
	StartTime  time.Time
	EndTime    time.Time
	Actor      string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionComplete) IsDashboardEvent() {}
