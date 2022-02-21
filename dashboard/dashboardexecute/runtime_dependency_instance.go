package dashboardexecute

import (
	"sync"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ResolvedRuntimeDependency is a wrapper for RuntimeDependency which contains the resolved value
// we must wrap it so that we do not mutate the underlying workdspace data when resolving dependency values
type ResolvedRuntimeDependency struct {
	dependency *modconfig.RuntimeDependency
	valueLock  sync.Mutex
	value      *string
}

func NewResolvedRuntimeDependency(dep *modconfig.RuntimeDependency) *ResolvedRuntimeDependency {
	return &ResolvedRuntimeDependency{dependency: dep}
}

func (d *ResolvedRuntimeDependency) IsResolved() bool {
	d.valueLock.Lock()
	defer d.valueLock.Unlock()
	return d.value != nil
}

func (d *ResolvedRuntimeDependency) Resolve() bool {
	d.valueLock.Lock()
	defer d.valueLock.Unlock()

	// did we succeed?
	d.value = d.dependency.SourceResource.GetValue()

	if d.value != nil {
		d.dependency.SetTargetFunc(*d.value)
		return true
	}

	return false
}
