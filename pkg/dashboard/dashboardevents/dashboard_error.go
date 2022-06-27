package dashboardevents

import "github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"

type DashboardError struct {
	Dashboard   dashboardtypes.DashboardNodeRun
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*DashboardError) IsDashboardEvent() {}
