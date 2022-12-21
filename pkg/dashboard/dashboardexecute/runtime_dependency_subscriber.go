package dashboardexecute

import "github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"

type RuntimeDependencySubscriber interface {
	RuntimeDependencyPublisher
	GetBaseDependencySubscriber() RuntimeDependencySubscriber
	GetRuntimeDependencyParent() dashboardtypes.DashboardParent
	SetRuntimeDependencyParent(dashboardtypes.DashboardParent)
}
