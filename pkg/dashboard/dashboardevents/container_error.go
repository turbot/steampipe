package dashboardevents

import "github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"

type ContainerError struct {
	Container   dashboardtypes.DashboardTreeRun
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ContainerError) IsDashboardEvent() {}
