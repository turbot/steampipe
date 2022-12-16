package dashboardexecute

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type RuntimeDependencySubscriberImpl struct {
	// all RuntimeDependencySubscribers are also publishers as they have args/params
	RuntimeDependencyPublisherImpl
	// map of runtime dependencies, keyed by dependency long name
	runtimeDependencies map[string]*dashboardtypes.ResolvedRuntimeDependency
}

func NewRuntimeDependencySubscriberImpl(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) RuntimeDependencySubscriberImpl {
	b := RuntimeDependencySubscriberImpl{
		RuntimeDependencyPublisherImpl: NewRuntimeDependencyPublisherImpl(resource, parent, executionTree),
		runtimeDependencies:            make(map[string]*dashboardtypes.ResolvedRuntimeDependency),
	}
	return b
}

func (b *RuntimeDependencySubscriberImpl) initRuntimeDependencies() error {
	// if the resource is a runtime dependency provider, create with runs and resolve dependencies
	rdp, ok := b.resource.(modconfig.RuntimeDependencyProvider)
	if !ok {
		return nil
	}
	// if we have with blocks, create runs for them
	// BEFORE creating child runs, and before adding runtime dependencies
	err := b.createWithRuns(rdp.GetWiths(), b.executionTree)
	if err != nil {
		return err
	}
	// resolve any runtime dependencies
	if err := b.resolveRuntimeDependencies(rdp); err != nil {
		return err
	}

	return nil
}

// if this node has runtime dependencies, find the publisher of the dependency and create a dashboardtypes.ResolvedRuntimeDependency
// which  we use to resolve the values
func (b *RuntimeDependencySubscriberImpl) resolveRuntimeDependencies(rdp modconfig.RuntimeDependencyProvider) error {
	runtimeDependencies := rdp.GetRuntimeDependencies()
	for n, d := range runtimeDependencies {
		// find a runtime dependency publisher who can provider this runtime dependency
		publisher := b.findRuntimeDependencyPublisher(d)
		if publisher == nil {
			// should never happen as validation should have caught this
			return fmt.Errorf("cannot resolve runtime dependency %s", d.String())
		}

		// read name and dep into local loop vars to ensure correct value used when transform func is invoked
		name := n
		dep := d

		// determine the function to use to retrieve the runtime dependency value
		var opts []RuntimeDependencyPublishOption

		switch dep.PropertyPath.ItemType {
		case modconfig.BlockTypeWith:
			// set a transform function to extract the requested with data
			opts = append(opts, WithTransform(func(resolvedVal *dashboardtypes.ResolvedRuntimeDependencyValue) *dashboardtypes.ResolvedRuntimeDependencyValue {
				transformedResolvedVal := &dashboardtypes.ResolvedRuntimeDependencyValue{Error: resolvedVal.Error}
				if resolvedVal.Error == nil {
					// the runtime dependency value for a 'with' is *dashboardtypes.LeafData
					withValue, err := b.getWithValue(name, resolvedVal.Value.(*dashboardtypes.LeafData), dep.PropertyPath)
					if err != nil {
						transformedResolvedVal.Error = fmt.Errorf("failed to resolve with value '%s' for %s: %s", dep.PropertyPath.Original, name, err.Error())
					} else {
						transformedResolvedVal.Value = withValue
					}
				}
				return transformedResolvedVal
			}))
		}
		// subscribe, passing a function which invokes getWithValue to resolve the required with value
		valueChannel := publisher.SubscribeToRuntimeDependency(d.SourceResourceName(), opts...)

		publisherName := publisher.GetName()
		b.runtimeDependencies[name] = dashboardtypes.NewResolvedRuntimeDependency(dep, valueChannel, publisherName)
	}
	return nil
}

func (b *RuntimeDependencySubscriberImpl) FindRuntimeDependenciesForParentProperty(parentProperty string) []*dashboardtypes.ResolvedRuntimeDependency {
	var res []*dashboardtypes.ResolvedRuntimeDependency
	for _, dep := range b.runtimeDependencies {
		if dep.Dependency.ParentPropertyName == parentProperty {
			res = append(res, dep)
		}
	}
	return res
}

func (b *RuntimeDependencySubscriberImpl) FindRuntimeDependencyForParentProperty(parentProperty string) *dashboardtypes.ResolvedRuntimeDependency {
	res := b.FindRuntimeDependenciesForParentProperty(parentProperty)
	if len(res) > 1 {
		panic(fmt.Sprintf("FindRuntimeDependencyForParentProperty for %s, parent property %s, returned more that 1 result", b.Name, parentProperty))
	}
	if res == nil {
		return nil
	}
	// return first result
	return res[0]
}
