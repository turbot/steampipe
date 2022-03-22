package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type DashboardError struct {
	Dashboard dashboardinterfaces.DashboardNodeRun
}

// IsDashboardEvent implements DashboardEvent interface
func (*DashboardError) IsDashboardEvent() {}
