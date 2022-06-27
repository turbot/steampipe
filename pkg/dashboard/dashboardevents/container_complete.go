package dashboardevents

import "github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"

type ContainerComplete struct {
	Container   dashboardtypes.DashboardNodeRun
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ContainerComplete) IsDashboardEvent() {}
