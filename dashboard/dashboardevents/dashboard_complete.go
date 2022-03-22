package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type DashboardComplete struct {
	Dashboard dashboardinterfaces.DashboardNodeRun
}

// IsDashboardEvent implements DashboardEvent interface
func (*DashboardComplete) IsDashboardEvent() {}
