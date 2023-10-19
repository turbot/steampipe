package dashboardexecute

import (
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
)

type RuntimeDependencyPublisher interface {
	dashboardtypes.DashboardTreeRun
	ProvidesRuntimeDependency(dependency *modconfig.RuntimeDependency) bool
	SubscribeToRuntimeDependency(name string, opts ...RuntimeDependencyPublishOption) chan *dashboardtypes.ResolvedRuntimeDependencyValue
	PublishRuntimeDependencyValue(name string, result *dashboardtypes.ResolvedRuntimeDependencyValue)
	GetWithRuns() map[string]*LeafRun
}
