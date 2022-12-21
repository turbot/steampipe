package dashboardexecute

import (
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
)

type LeafRunOption = func(target *LeafRun)

func setName(name string) LeafRunOption {
	return func(target *LeafRun) {
		target.Name = name
	}
}
func setRuntimeDependencyParent(parent dashboardtypes.DashboardParent) LeafRunOption {
	return func(target *LeafRun) {
		target.SetRuntimeDependencyParent(parent)
	}
}
