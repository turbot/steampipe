package dashboardevents

import (
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"time"
)

type ExecutionStarted struct {
	Root        dashboardtypes.DashboardTreeRun `json:"dashboard"`
	Panels      map[string]any
	Session     string
	ExecutionId string
	Inputs      map[string]any
	Variables   map[string]string
	StartTime   time.Time
	// immutable representation of event data - to avoid mutation before we send it
	JsonData []byte
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionStarted) IsDashboardEvent() {}
