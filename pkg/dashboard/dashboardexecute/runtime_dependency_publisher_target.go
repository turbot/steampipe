package dashboardexecute

import (
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
)

type RuntimeDependencyPublishTarget struct {
	transform func(*dashboardtypes.ResolvedRuntimeDependencyValue) *dashboardtypes.ResolvedRuntimeDependencyValue
	channel   chan *dashboardtypes.ResolvedRuntimeDependencyValue
}
