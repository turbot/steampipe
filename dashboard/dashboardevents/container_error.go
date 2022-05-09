package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type ContainerError struct {
	Container   dashboardinterfaces.DashboardNodeRun
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ContainerError) IsDashboardEvent() {}
