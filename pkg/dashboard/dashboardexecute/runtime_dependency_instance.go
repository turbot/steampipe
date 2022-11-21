package dashboardexecute

import (
	"sync"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// ResolvedRuntimeDependency is a wrapper for RuntimeDependency which contains the resolved value
// we must wrap it so that we do not mutate the underlying workspace data when resolving dependency values
type ResolvedRuntimeDependency struct {
	dependency   *modconfig.RuntimeDependency
	valueLock    sync.Mutex
	value        any
	getValueFunc func(string) (any, error)
}

func NewResolvedRuntimeDependency(dep *modconfig.RuntimeDependency, getValueFunc func(string) (any, error)) *ResolvedRuntimeDependency {
	return &ResolvedRuntimeDependency{
		dependency:   dep,
		getValueFunc: getValueFunc,
	}
}

func (d *ResolvedRuntimeDependency) Resolve() (bool, error) {
	d.valueLock.Lock()
	defer d.valueLock.Unlock()

	// if we are already resolved, do nothing
	if d.hasValue() {
		return true, nil
	}

	// otherwise, try to read the value from the source
	val, err := d.getValueFunc(d.dependency.SourceResource.GetUnqualifiedName())
	if err != nil {
		return false, err
	}
	d.value = val

	// did we succeed
	return d.hasValue(), nil
}

func (d *ResolvedRuntimeDependency) hasValue() bool {
	return !helpers.IsNil(d.value)
}
