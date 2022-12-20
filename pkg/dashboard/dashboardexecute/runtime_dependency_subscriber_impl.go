package dashboardexecute

import (
	"context"
	"fmt"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"golang.org/x/exp/maps"
	"log"
)

type RuntimeDependencySubscriber struct {
	// all RuntimeDependencySubscribers are also publishers as they have args/params
	RuntimeDependencyPublisherImpl
	// map of runtime dependencies, keyed by dependency long name
	runtimeDependencies map[string]*dashboardtypes.ResolvedRuntimeDependency
	// a list of the (scoped) names of any runtime dependencies that we rely on
	RuntimeDependencyNames   []string `json:"dependencies,omitempty"`
	RawSQL                   string   `json:"sql,omitempty"`
	executeSQL               string
	baseDependencySubscriber *RuntimeDependencySubscriber
}

func NewRuntimeDependencySubscriber(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, run dashboardtypes.DashboardTreeRun, executionTree *DashboardExecutionTree) *RuntimeDependencySubscriber {
	b := &RuntimeDependencySubscriber{
		runtimeDependencies: make(map[string]*dashboardtypes.ResolvedRuntimeDependency),
	}
	// TODO [node_reuse]
	// HACK
	// if this is a run for a base resource there will be no 'run'
	if run == nil {
		run = b
	}

	// create RuntimeDependencyPublisherImpl
	// (we must create after creating the run as iut requires a ref to the run)
	b.RuntimeDependencyPublisherImpl = NewRuntimeDependencyPublisherImpl(resource, parent, run, executionTree)

	return b
}

// if the resource is a runtime dependency provider, create with runs and resolve dependencies
func (s *RuntimeDependencySubscriber) initRuntimeDependencies(executionTree *DashboardExecutionTree) error {
	if _, ok := s.resource.(modconfig.RuntimeDependencyProvider); !ok {
		return nil
	}

	// if our underlying resource has a base which has runtime dependencies,
	// create a RuntimeDependencySubscriber for it
	if err := s.initBaseRuntimeDependencySubscriber(executionTree); err != nil {
		return err
	}

	// call into publisher to start any with runs
	if err := s.RuntimeDependencyPublisherImpl.initRuntimeDependencies(); err != nil {
		return err
	}
	// resolve any runtime dependencies
	return s.resolveRuntimeDependencies()
}

func (s *RuntimeDependencySubscriber) initBaseRuntimeDependencySubscriber(executionTree *DashboardExecutionTree) error {
	if base := s.resource.(modconfig.HclResource).GetBase(); base != nil {
		if _, ok := base.(modconfig.RuntimeDependencyProvider); ok {
			s.baseDependencySubscriber = NewRuntimeDependencySubscriber(base.(modconfig.DashboardLeafNode), nil, s, executionTree)
			err := s.baseDependencySubscriber.initRuntimeDependencies(executionTree)
			if err != nil {
				return err
			}
			// create buffered channel for base with to report their completion
			s.baseDependencySubscriber.createChildCompleteChan()
		}
	}
	return nil
}

// if this node has runtime dependencies, find the publisher of the dependency and create a dashboardtypes.ResolvedRuntimeDependency
// which  we use to resolve the values
func (s *RuntimeDependencySubscriber) resolveRuntimeDependencies() error {
	rdp, ok := s.resource.(modconfig.RuntimeDependencyProvider)
	if !ok {
		return nil
	}

	runtimeDependencies := rdp.GetRuntimeDependencies()

	for n, d := range runtimeDependencies {
		// find a runtime dependency publisher who can provider this runtime dependency
		publisher := s.findRuntimeDependencyPublisher(d)
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
					withValue, err := s.getWithValue(name, resolvedVal.Value.(*dashboardtypes.LeafData), dep.PropertyPath)
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
		s.runtimeDependencies[name] = dashboardtypes.NewResolvedRuntimeDependency(dep, valueChannel, publisherName)
	}

	return nil
}

func (s *RuntimeDependencySubscriber) findRuntimeDependencyPublisher(runtimeDependency *modconfig.RuntimeDependency) (res RuntimeDependencyPublisher) {
	// the runtime dependency publisher is usually the root node of the execution tree
	// the exception to this is if the node is a LeafRun(?) for a base node which has a with block,
	// in which case it may provide its own runtime dependency

	// traverse up the tree - we rely on resource validation to ensure that intermediate nodes in the tree
	// do not provide the runtime dependency
	var item dashboardtypes.DashboardTreeRun = s
	for {

		if publisher, ok := item.(RuntimeDependencyPublisher); ok {
			if publisher.ProvidesRuntimeDependency(runtimeDependency) {
				res = publisher
				break
			}
		}

		item = item.GetParent()
		if item == s.executionTree {
			break
		}
	}
	return
}

func (s *RuntimeDependencySubscriber) evaluateRuntimeDependencies() error {
	// now wait for any runtime dependencies then resolve args and params
	// (it is possible to have params but no sql)
	if s.hasRuntimeDependencies() {
		// if there are any unresolved runtime dependencies, wait for them
		if err := s.waitForRuntimeDependencies(); err != nil {
			return err
		}

		// ok now we have runtime dependencies, we can resolve the query
		if err := s.resolveSQLAndArgs(); err != nil {
			return err
		}
	}
	return nil
}

func (s *RuntimeDependencySubscriber) waitForRuntimeDependencies() error {
	if !s.hasRuntimeDependencies() {
		return nil
	}

	// wait for base dependencies if we have any
	if s.baseDependencySubscriber != nil {
		if err := s.baseDependencySubscriber.waitForRuntimeDependencies(); err != nil {
			return err
		}
	}

	allRuntimeDepsResolved := true
	for _, dep := range s.runtimeDependencies {
		if !dep.IsResolved() {
			allRuntimeDepsResolved = false
		}
	}
	if allRuntimeDepsResolved {
		return nil
	}

	// set status to blocked
	s.setStatus(dashboardtypes.DashboardRunBlocked)

	log.Printf("[TRACE] LeafRun '%s' waitForRuntimeDependencies", s.resource.Name())
	for _, resolvedDependency := range s.runtimeDependencies {
		// TODO [node_reuse] what about dependencies _between_ dependencies - do this async
		// block until the dependency is available
		err := resolvedDependency.Resolve()
		if err != nil {
			return err
		}
	}

	log.Printf("[TRACE] %s: all runtime dependencies ready", s.resource.Name())
	return nil
}

func (s *RuntimeDependencySubscriber) findRuntimeDependenciesForParentProperty(parentProperty string) []*dashboardtypes.ResolvedRuntimeDependency {
	var res []*dashboardtypes.ResolvedRuntimeDependency
	for _, dep := range s.runtimeDependencies {
		if dep.Dependency.ParentPropertyName == parentProperty {
			res = append(res, dep)
		}
	}
	// also look at base subscriber
	if s.baseDependencySubscriber != nil {
		for _, dep := range s.baseDependencySubscriber.runtimeDependencies {
			if dep.Dependency.ParentPropertyName == parentProperty {
				res = append(res, dep)
			}
		}
	}
	return res
}

func (s *RuntimeDependencySubscriber) findRuntimeDependencyForParentProperty(parentProperty string) *dashboardtypes.ResolvedRuntimeDependency {
	res := s.findRuntimeDependenciesForParentProperty(parentProperty)
	if len(res) > 1 {
		panic(fmt.Sprintf("findRuntimeDependencyForParentProperty for %s, parent property %s, returned more that 1 result", s.Name, parentProperty))
	}
	if res == nil {
		return nil
	}
	// return first result
	return res[0]
}

// resolve the sql for this leaf run into the source sql (i.e. NOT the prepared statement name) and resolved args
func (s *RuntimeDependencySubscriber) resolveSQLAndArgs() error {
	log.Printf("[TRACE] LeafRun '%s' resolveSQLAndArgs", s.resource.Name())
	queryProvider, ok := s.resource.(modconfig.QueryProvider)
	if !ok {
		// not a query provider - nothing to do
		return nil
	}

	// convert arg runtime dependencies into arg map
	runtimeArgs, err := s.buildRuntimeDependencyArgs()
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs failed: %s", s.resource.Name(), err.Error())
		return err
	}

	// now if any param defaults had runtime dependencies, populate them
	s.populateParamDefaults(queryProvider)

	log.Printf("[TRACE] LeafRun '%s' built runtime args: %v", s.resource.Name(), runtimeArgs)

	// does this leaf run have any SQL to execute?
	// TODO [node_reuse] split this into resolve query and resolve args - we may have args but no query
	if queryProvider.RequiresExecution(queryProvider) {
		resolvedQuery, err := s.executionTree.workspace.ResolveQueryFromQueryProvider(queryProvider, runtimeArgs)
		if err != nil {
			return err
		}
		s.RawSQL = resolvedQuery.RawSQL
		s.executeSQL = resolvedQuery.ExecuteSQL
		s.Args = resolvedQuery.Args
	}
	//}
	return nil
}

func (s *RuntimeDependencySubscriber) populateParamDefaults(provider modconfig.QueryProvider) {
	paramDefs := provider.GetParams()
	for _, paramDef := range paramDefs {
		if dep := s.findRuntimeDependencyForParentProperty(paramDef.UnqualifiedName); dep != nil {
			// assuming the default property is the target, set the default
			if typehelpers.SafeString(dep.Dependency.TargetPropertyName) == "default" {
				paramDef.SetDefault(dep.Value)
			}
		}
	}
}

// convert runtime dependencies into arg map
func (s *RuntimeDependencySubscriber) buildRuntimeDependencyArgs() (*modconfig.QueryArgs, error) {
	res := modconfig.NewQueryArgs()

	log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs - %d runtime dependencies", s.resource.Name(), len(s.runtimeDependencies))

	// if the runtime dependencies use position args, get the max index and ensure the args array is large enough
	maxArgIndex := -1
	// build list of all args runtime dependencies
	argRuntimeDependencies := s.findRuntimeDependenciesForParentProperty(modconfig.AttributeArgs)

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

func (s *RuntimeDependencySubscriber) hasParam(paramName string) bool {
	for _, p := range s.Params {
		if p.ShortName == paramName {
			return true
		}
	}
	return false
}

// populate the list of runtime dependencies that this run depends on
func (s *RuntimeDependencySubscriber) setRuntimeDependencies() {
	names := make(map[string]struct{}, len(s.runtimeDependencies))
	for _, d := range s.runtimeDependencies {
		// add to DependencyWiths using ScopedName, i.e. <parent FullName>.<with UnqualifiedName>.
		// we do this as there may be a with from a base resource with a clashing with name
		// NOTE: this must be consistent with the naming in RuntimeDependencyPublisherImpl.createWithRuns
		names[d.ScopedName()] = struct{}{}
	}

	// get base runtime dependencies (if any)
	if s.baseDependencySubscriber != nil {
		s.baseDependencySubscriber.setRuntimeDependencies()
		s.RuntimeDependencyNames = append(s.RuntimeDependencyNames, s.baseDependencySubscriber.RuntimeDependencyNames...)
	}
	s.RuntimeDependencyNames = maps.Keys(names)
}

func (s *RuntimeDependencySubscriber) hasRuntimeDependencies() bool {
	return len(s.runtimeDependencies)+len(s.baseRuntimeDependencies()) > 0
}

func (s *RuntimeDependencySubscriber) baseRuntimeDependencies() map[string]*dashboardtypes.ResolvedRuntimeDependency {
	if s.baseDependencySubscriber == nil {
		return map[string]*dashboardtypes.ResolvedRuntimeDependency{}
	}
	return s.baseDependencySubscriber.runtimeDependencies
}

// override DashboardParentImpl.executeChildrenAsync to also execute 'withs' of our baseRun
func (s *RuntimeDependencySubscriber) executeChildrenAsync(ctx context.Context) {
	// if we have a baseDependencySubscriber, execute it
	if s.baseDependencySubscriber != nil {
		go s.baseDependencySubscriber.executeWithsAsync(ctx)
	}

	// if this leaf run has children (including with runs) execute them asynchronously

	// set RuntimeDependenciesOnly if needed
	s.DashboardParentImpl.executeChildrenAsync(ctx)
}
