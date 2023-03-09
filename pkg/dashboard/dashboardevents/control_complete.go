package dashboardevents

import (
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"time"
)

type ControlComplete struct {
	Progress    *controlstatus.ControlProgress
	Control     controlstatus.ControlRunStatusProvider
	Name        string
	Session     string
	ExecutionId string
	Timestamp   time.Time
}

// IsDashboardEvent implements DashboardEvent interface
func (*ControlComplete) IsDashboardEvent() {}
