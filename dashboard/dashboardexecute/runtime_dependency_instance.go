package dashboardexecute

import (
	"sync"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ResolvedRuntimeDependency is a wrapper for RuntimeDependency which contains the resolved value
// we must wrap it so that we do not mutate the underlying workspace data when resolving dependency values
type ResolvedRuntimeDependency struct {
	dependency *modconfig.RuntimeDependency
	valueLock  sync.Mutex
	value      *string
}

func NewResolvedRuntimeDependency(dep *modconfig.RuntimeDependency) *ResolvedRuntimeDependency {
	return &ResolvedRuntimeDependency{dependency: dep}
}

func (d *ResolvedRuntimeDependency) Resolve() bool {
	d.valueLock.Lock()
	defer d.valueLock.Unlock()

	// if we are already resolved, do nothing
	if d.value != nil {
		return true
	}

	// otherwise try to read the value from the source
	d.value = d.dependency.SourceResource.GetValue()

	// did we succeed
	if d.value != nil {
		// if so, set the target property
		d.dependency.SetTargetFunc(*d.value)
		return true
	}

	return false
}
