package dashboardexecute

import (
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type RuntimeDependencyPublisher interface {
	GetName() string
	ProvidesRuntimeDependency(dependency *modconfig.RuntimeDependency) bool
	SubscribeToRuntimeDependency(name string, opts ...RuntimeDependencyPublishOption) chan *dashboardtypes.ResolvedRuntimeDependencyValue
	PublishRuntimeDependencyValue(name string, result *dashboardtypes.ResolvedRuntimeDependencyValue)
	GetParentPublisher() RuntimeDependencyPublisher
	GetWithRuns() map[string]*LeafRun
}
