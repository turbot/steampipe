package dashboardexecute

import (
	"context"
	"fmt"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"log"
)

type RuntimeDependencySubscriberImpl struct {
	// all RuntimeDependencySubscribers are also publishers as they have args/params
	RuntimeDependencyPublisherImpl
	// map of runtime dependencies, keyed by dependency long name
	runtimeDependencies map[string]*dashboardtypes.ResolvedRuntimeDependency
	// a list of the (scoped) names of any runtime dependencies that we rely on
	RuntimeDependencyNames []string `json:"dependencies,omitempty"`
	RawSQL                 string   `json:"sql,omitempty"`
	executeSQL             string
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

	if err := b.RuntimeDependencyPublisherImpl.initRuntimeDependencies(); err != nil {
		return err
	}
	// resolve any runtime dependencies
	return b.resolveRuntimeDependencies(rdp)
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

func (b *RuntimeDependencySubscriberImpl) evaluateRuntimeDependencies(ctx context.Context) error {
	// now wait for any runtime dependencies then resolve args and params
	// (it is possible to have params but no sql)
	if len(b.runtimeDependencies) > 0 {
		// if there are any unresolved runtime dependencies, wait for them
		if err := b.waitForRuntimeDependencies(); err != nil {

			return err
		}

		// ok now we have runtime dependencies, we can resolve the query
		if err := b.resolveSQLAndArgs(); err != nil {
			return err
		}
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

func (b *RuntimeDependencySubscriberImpl) waitForRuntimeDependencies() error {
	allRuntimeDepsResolved := true
	for _, dep := range b.runtimeDependencies {
		if !dep.IsResolved() {
			allRuntimeDepsResolved = false
		}
	}
	if allRuntimeDepsResolved {
		return nil
	}

	// set status to blocked
	b.setStatus(dashboardtypes.DashboardRunBlocked)

	log.Printf("[TRACE] LeafRun '%s' waitForRuntimeDependencies", b.resource.Name())
	for _, resolvedDependency := range b.runtimeDependencies {
		// check whether the dependency is available
		err := resolvedDependency.Resolve()
		if err != nil {
			return err
		}
	}

	if len(b.runtimeDependencies) > 0 {
		log.Printf("[TRACE] LeafRun '%s' all runtime dependencies ready", b.resource.Name())
	}
	return nil
}

// resolve the sql for this leaf run into the source sql (i.e. NOT the prepared statement name) and resolved args
func (b *RuntimeDependencySubscriberImpl) resolveSQLAndArgs() error {
	log.Printf("[TRACE] LeafRun '%s' resolveSQLAndArgs", b.resource.Name())
	queryProvider, ok := b.resource.(modconfig.QueryProvider)
	if !ok {
		// not a query provider - nothing to do
		return nil
	}

	// convert arg runtime dependencies into arg map
	runtimeArgs, err := b.buildRuntimeDependencyArgs()
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs failed: %s", b.resource.Name(), err.Error())
		return err
	}

	// now if any param defaults had runtime dependencies, populate them
	b.populateParamDefaults(queryProvider)

	log.Printf("[TRACE] LeafRun '%s' built runtime args: %v", b.resource.Name(), runtimeArgs)

	// does this leaf run have any SQL to execute?
	// TODO [node_reuse] split this into resolve query and resolve args - we may have args but no query
	if queryProvider.RequiresExecution(queryProvider) {
		resolvedQuery, err := b.executionTree.workspace.ResolveQueryFromQueryProvider(queryProvider, runtimeArgs)
		if err != nil {
			return err
		}
		b.RawSQL = resolvedQuery.RawSQL
		b.executeSQL = resolvedQuery.ExecuteSQL
		b.Args = resolvedQuery.Args
	}
	//}
	return nil
}
func (b *RuntimeDependencySubscriberImpl) populateParamDefaults(provider modconfig.QueryProvider) {
	paramDefs := provider.GetParams()
	for _, paramDef := range paramDefs {
		if dep := b.FindRuntimeDependencyForParentProperty(paramDef.UnqualifiedName); dep != nil {
			// assuming the default property is the target, set the default
			if typehelpers.SafeString(dep.Dependency.TargetPropertyName) == "default" {
				paramDef.SetDefault(dep.Value)
			}
		}
	}
}

// convert runtime dependencies into arg map
func (b *RuntimeDependencySubscriberImpl) buildRuntimeDependencyArgs() (*modconfig.QueryArgs, error) {
	res := modconfig.NewQueryArgs()

	log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs - %d runtime dependencies", b.resource.Name(), len(b.runtimeDependencies))

	// if the runtime dependencies use position args, get the max index and ensure the args array is large enough
	maxArgIndex := -1
	// build list of all args runtime dependencies
	argRuntimeDependencies := b.FindRuntimeDependenciesForParentProperty(modconfig.AttributeArgs)

	for _, dep := range argRuntimeDependencies {
		if dep.Dependency.TargetPropertyIndex != nil && *dep.Dependency.TargetPropertyIndex > maxArgIndex {
			maxArgIndex = *dep.Dependency.TargetPropertyIndex
		}
	}
	if maxArgIndex != -1 {
		res.ArgList = make([]*string, maxArgIndex+1)
	}

	// now set the arg values
	for _, dep := range argRuntimeDependencies {
		if dep.Dependency.TargetPropertyName != nil {
			err := res.SetNamedArgVal(dep.Value, *dep.Dependency.TargetPropertyName)
			if err != nil {
				return nil, err
			}

		} else {
			if dep.Dependency.TargetPropertyIndex == nil {
				return nil, fmt.Errorf("invalid runtime dependency - both ArgName and ArgIndex are nil ")
			}
			err := res.SetPositionalArgVal(dep.Value, *dep.Dependency.TargetPropertyIndex)
			if err != nil {
				return nil, err
			}
		}
	}
	return res, nil
}

func (b *RuntimeDependencySubscriberImpl) hasParam(paramName string) bool {
	for _, p := range b.Params {
		if p.ShortName == paramName {
			return true
		}
	}
	return false
}

// populate the list of runtime dependencies that this run depends on
func (r *RuntimeDependencySubscriberImpl) setRuntimeDependencies() {
	for _, d := range r.runtimeDependencies {
		// add to DependencyWiths using ScopedName, i.e. <parent FullName>.<with UnqualifiedName>.
		// we do this as there may be a with from a base resource with a clashing with name
		// NOTE: this must be consistent with the naming in RuntimeDependencyPublisherImpl.createWithRuns
		r.RuntimeDependencyNames = append(r.RuntimeDependencyNames, d.ScopedName())
	}
}
