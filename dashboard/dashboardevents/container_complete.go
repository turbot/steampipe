package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type ContainerComplete struct {
	Container dashboardinterfaces.DashboardNodeRun
}

// IsDashboardEvent implements DashboardEvent interface
func (*ContainerComplete) IsDashboardEvent() {}
