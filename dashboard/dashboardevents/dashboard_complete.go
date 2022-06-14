package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardtypes"

type DashboardComplete struct {
	Dashboard dashboardtypes.DashboardNodeRun
}

// IsDashboardEvent implements DashboardEvent interface
func (*DashboardComplete) IsDashboardEvent() {}
