package dashboardtypes

import (
	"fmt"
	"github.com/turbot/go-kit/type_conversion"
	"log"
	"sync"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
)

// ResolvedRuntimeDependency is a wrapper for RuntimeDependency which contains the resolved value
// we must wrap it so that we do not mutate the underlying workspace data when resolving dependency values
type ResolvedRuntimeDependency struct {
	Dependency *modconfig.RuntimeDependency
	valueLock  sync.Mutex
	Value      any
	// the name of the run which publishes this dependency
	publisherName string
	valueChannel  chan *ResolvedRuntimeDependencyValue
}

func NewResolvedRuntimeDependency(dep *modconfig.RuntimeDependency, valueChannel chan *ResolvedRuntimeDependencyValue, publisherName string) *ResolvedRuntimeDependency {
	return &ResolvedRuntimeDependency{
		Dependency:    dep,
		valueChannel:  valueChannel,
		publisherName: publisherName,
	}
}

// ScopedName returns is a unique name for the dependency by prepending the publisher name
// this is used to uniquely identify which `with` is used - for the snapshot data
func (d *ResolvedRuntimeDependency) ScopedName() string {
	return fmt.Sprintf("%s.%s", d.publisherName, d.Dependency.SourceResourceName())
}

func (d *ResolvedRuntimeDependency) IsResolved() bool {
	d.valueLock.Lock()
	defer d.valueLock.Unlock()

	return d.hasValue()
}

func (d *ResolvedRuntimeDependency) Resolve() error {
	d.valueLock.Lock()
	defer d.valueLock.Unlock()

	log.Printf("[TRACE] ResolvedRuntimeDependency Resolve dep %s chan %p", d.Dependency.PropertyPath, d.valueChannel)

	// if we are already resolved, do nothing
	if d.hasValue() {
		return nil
	}

	// wait for value
	val := <-d.valueChannel

	d.Value = val.Value

	// TACTICAL if the desired value is an array, wrap in an array
	if d.Dependency.IsArray {
		d.Value = type_conversion.AnySliceToTypedSlice([]any{d.Value})
	}

	if val.Error != nil {
		return val.Error
	}

	// we should have a non nil value now
	if !d.hasValue() {
		return fmt.Errorf("nil value recevied for runtime dependency %s", d.Dependency.String())
	}
	return nil
}

func (d *ResolvedRuntimeDependency) hasValue() bool {
	return !helpers.IsNil(d.Value)
}
