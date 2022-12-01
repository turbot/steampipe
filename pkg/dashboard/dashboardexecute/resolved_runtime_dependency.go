package dashboardexecute

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"log"
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
	valueChannel chan *dashboardtypes.ResolvedRuntimeDependencyValue
}

func NewResolvedRuntimeDependency(dep *modconfig.RuntimeDependency, valueChannel chan *dashboardtypes.ResolvedRuntimeDependencyValue) *ResolvedRuntimeDependency {
	var wg sync.WaitGroup
	wg.Add(1)
	return &ResolvedRuntimeDependency{
		dependency:   dep,
		valueChannel: valueChannel,
	}
}

func (d *ResolvedRuntimeDependency) Resolve() error {
	d.valueLock.Lock()
	defer d.valueLock.Unlock()

	log.Printf("[TRACE] ResolvedRuntimeDependency Resolve dep %s chan %p", d.dependency.PropertyPath, d.valueChannel)

	// if we are already resolved, do nothing
	if d.hasValue() {
		return nil
	}

	// dependency must have a source resource - if not this is a bug
	if d.dependency.SourceResource == nil {
		return fmt.Errorf("runtime dependency '%s' Resolve() called but it does not have a source resource", d.dependency.String())
	}

	// wait for value
	val := <-d.valueChannel

	d.value = val.Value
	// TODO will this be reported OK?
	if val.Error != nil {
		return val.Error
	}

	// we should have a non nil value now
	if !d.hasValue() {
		return fmt.Errorf("nil value recevied for runtime dependency %s", d.dependency.String())
	}
	return nil
}

func (d *ResolvedRuntimeDependency) hasValue() bool {
	return !helpers.IsNil(d.value)
}
