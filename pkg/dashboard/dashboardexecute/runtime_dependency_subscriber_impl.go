package dashboardexecute

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"golang.org/x/exp/maps"
)

type RuntimeDependencySubscriberImpl struct {
	// all RuntimeDependencySubscribers are also publishers as they have args/params
	runtimeDependencyPublisherImpl
	// if the underlying resource has a base resource, create a RuntimeDependencySubscriberImpl instance to handle
	// generation and publication of runtime depdencies from the base resource
	baseDependencySubscriber *RuntimeDependencySubscriberImpl
	// map of runtime dependencies, keyed by dependency long name
	runtimeDependencies map[string]*dashboardtypes.ResolvedRuntimeDependency
	RawSQL              string `json:"sql,omitempty"`
	executeSQL          string
	// a list of the (scoped) names of any runtime dependencies that we rely on
	RuntimeDependencyNames []string `json:"dependencies,omitempty"`
}

func NewRuntimeDependencySubscriber(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, run dashboardtypes.DashboardTreeRun, executionTree *DashboardExecutionTree) *RuntimeDependencySubscriberImpl {
	b := &RuntimeDependencySubscriberImpl{
		runtimeDependencies: make(map[string]*dashboardtypes.ResolvedRuntimeDependency),
	}

	// create RuntimeDependencyPublisherImpl
	// (we must create after creating the run as iut requires a ref to the run)
	b.runtimeDependencyPublisherImpl = newRuntimeDependencyPublisherImpl(resource, parent, run, executionTree)

	return b
}

// GetBaseDependencySubscriber implements RuntimeDependencySubscriber
func (s *RuntimeDependencySubscriberImpl) GetBaseDependencySubscriber() RuntimeDependencySubscriber {
	return s.baseDependencySubscriber
}

// if the resource is a runtime dependency provider, create with runs and resolve dependencies
func (s *RuntimeDependencySubscriberImpl) initRuntimeDependencies(executionTree *DashboardExecutionTree) error {
	if _, ok := s.resource.(modconfig.RuntimeDependencyProvider); !ok {
		return nil
	}

	// if our underlying resource has a base which has runtime dependencies,
	// create a RuntimeDependencySubscriberImpl for it
	if err := s.initBaseRuntimeDependencySubscriber(executionTree); err != nil {
		return err
	}

	// call into publisher to start any with runs
	if err := s.runtimeDependencyPublisherImpl.initWiths(); err != nil {
		return err
	}
	// resolve any runtime dependencies
	return s.resolveRuntimeDependencies()
}

func (s *RuntimeDependencySubscriberImpl) initBaseRuntimeDependencySubscriber(executionTree *DashboardExecutionTree) error {
	if base := s.resource.(modconfig.HclResource).GetBase(); base != nil {
		if _, ok := base.(modconfig.RuntimeDependencyProvider); ok {
			// create base dependency subscriber
			// pass ourselves as 'run'
			// - this is only used when sending update events, which will not happen for the baseDependencySubscriber
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
func (s *RuntimeDependencySubscriberImpl) resolveRuntimeDependencies() error {
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

func (s *RuntimeDependencySubscriberImpl) findRuntimeDependencyPublisher(runtimeDependency *modconfig.RuntimeDependency) RuntimeDependencyPublisher {
	// the runtime dependency publisher is either the root dashboard run,
	// or if this resource (or in case of a node/edge, the resource parent) has a base,
	// the baseDependencySubscriber for that base
	var subscriber RuntimeDependencySubscriber = s
	if s.NodeType == modconfig.BlockTypeNode || s.NodeType == modconfig.BlockTypeEdge {
		subscriber = s.parent.(RuntimeDependencySubscriber)
	}
	baseSubscriber := subscriber.GetBaseDependencySubscriber()

	// "if I have a base with runtime dependencies, those dependencies must be provided BY THE BASE"
	// check the provider property on the runtime dependency
	// - if the matches the underlying resource for the baseDependencySubscriber,
	// then baseDependencySubscriber _should_ be the dependency publisher
	if !helpers.IsNil(baseSubscriber) && runtimeDependency.Provider == baseSubscriber.GetResource() {
		if baseSubscriber.ProvidesRuntimeDependency(runtimeDependency) {
			return baseSubscriber
		}

		// unexpected
		log.Printf("[WARN] dependency %s has a dependency provider matching the base resource %s but the BaseDependencySubscriber does not provider the runtime dependency",
			runtimeDependency.String(), baseSubscriber.GetName())
		return nil
	}

	// "if I am a base resource with runtime dependencies, I provide my own dependencies"
	// see if we can satisfy the dependency (this would occur when initialising the baseDependencySubscriber)
	if s.ProvidesRuntimeDependency(runtimeDependency) {
		return s
	}

	// "if I am a nested resource, my dashboard provides my dependencies"
	// otherwise the dashboard run must be the publisher
	dashboardRun := s.executionTree.runs[s.DashboardName].(RuntimeDependencyPublisher)
	if dashboardRun.ProvidesRuntimeDependency(runtimeDependency) {
		return dashboardRun
	}

	return nil
}

func (s *RuntimeDependencySubscriberImpl) evaluateRuntimeDependencies(ctx context.Context) error {
	log.Printf("[TRACE] %s: evaluateRuntimeDependencies", s.Name)
	// now wait for any runtime dependencies then resolve args and params
	// (it is possible to have params but no sql)
	if s.hasRuntimeDependencies() {
		// if there are any unresolved runtime dependencies, wait for them
		if err := s.waitForRuntimeDependencies(ctx); err != nil {
			return err
		}
		log.Printf("[TRACE] %s: runtime dependencies availablem resolving sql and args", s.Name)

		// ok now we have runtime dependencies, we can resolve the query
		if err := s.resolveSQLAndArgs(); err != nil {
			return err
		}
		// call the argsResolved callback in case anyone is waiting for the args
		s.argsResolved(s.Args)
	}
	return nil
}

func (s *RuntimeDependencySubscriberImpl) waitForRuntimeDependencies(ctx context.Context) error {
	log.Printf("[TRACE] %s: waitForRuntimeDependencies", s.Name)

	if !s.hasRuntimeDependencies() {
		log.Printf("[TRACE] %s: no runtime dependencies", s.Name)
		return nil
	}

	// wait for base dependencies if we have any
	if s.baseDependencySubscriber != nil {
		log.Printf("[TRACE] %s: calling baseDependencySubscriber.waitForRuntimeDependencies", s.Name)
		if err := s.baseDependencySubscriber.waitForRuntimeDependencies(ctx); err != nil {
			return err
		}
	}

	log.Printf("[TRACE] %s: checking whether all depdencies are resolved", s.Name)

	allRuntimeDepsResolved := true
	for _, dep := range s.runtimeDependencies {
		if !dep.IsResolved() {
			allRuntimeDepsResolved = false
			log.Printf("[TRACE] %s:  dependency %s is NOT resolved", s.Name, dep.Dependency.String())
		}
	}
	if allRuntimeDepsResolved {
		return nil
	}

	log.Printf("[TRACE] %s: BLOCKED", s.Name)
	// set status to blocked
	s.setStatus(ctx, dashboardtypes.RunBlocked)

	var wg sync.WaitGroup
	var errChan = make(chan error)
	var doneChan = make(chan struct{})
	for _, r := range s.runtimeDependencies {
		if !r.IsResolved() {
			// make copy of loop var for goroutine
			resolvedDependency := r
			log.Printf("[TRACE] %s: wait for %s", s.Name, resolvedDependency.Dependency.String())
			wg.Add(1)
			go func() {
				defer wg.Done()
				// block until the dependency is available
				err := resolvedDependency.Resolve()
				log.Printf("[TRACE] %s: Resolve returned for %s", s.Name, resolvedDependency.Dependency.String())
				if err != nil {
					log.Printf("[TRACE] %s: Resolve for %s returned error:L %s", s.Name, resolvedDependency.Dependency.String(), err.Error())
					errChan <- err
				}
			}()
		}
	}
	go func() {
		log.Printf("[TRACE] %s: goroutine waiting for all runtime deps to be available", s.Name)
		wg.Wait()
		close(doneChan)
	}()

	var errors []error

wait_loop:
	for {
		select {
		case err := <-errChan:
			errors = append(errors, err)
		case <-doneChan:
			break wait_loop
		case <-ctx.Done():
			errors = append(errors, ctx.Err())
			break wait_loop
		}
	}

	log.Printf("[TRACE] %s: all runtime dependencies ready", s.resource.Name())
	return error_helpers.CombineErrors(errors...)
}

func (s *RuntimeDependencySubscriberImpl) findRuntimeDependenciesForParentProperty(parentProperty string) []*dashboardtypes.ResolvedRuntimeDependency {
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

func (s *RuntimeDependencySubscriberImpl) findRuntimeDependencyForParentProperty(parentProperty string) *dashboardtypes.ResolvedRuntimeDependency {
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

// resolve the sql for this leaf run into the source sql and resolved args
func (s *RuntimeDependencySubscriberImpl) resolveSQLAndArgs() error {
	log.Printf("[TRACE] %s: resolveSQLAndArgs", s.resource.Name())
	queryProvider, ok := s.resource.(modconfig.QueryProvider)
	if !ok {
		// not a query provider - nothing to do
		return nil
	}

	// convert arg runtime dependencies into arg map
	runtimeArgs, err := s.buildRuntimeDependencyArgs()
	if err != nil {
		log.Printf("[TRACE] %s: buildRuntimeDependencyArgs failed: %s", s.resource.Name(), err.Error())
		return err
	}

	// now if any param defaults had runtime dependencies, populate them
	s.populateParamDefaults(queryProvider)

	log.Printf("[TRACE] %s: built runtime args: %v", s.resource.Name(), runtimeArgs)

	// does this leaf run have any SQL to execute?
	if queryProvider.RequiresExecution(queryProvider) {
		log.Printf("[TRACE] ResolveArgsFromQueryProvider for %s", queryProvider.Name())
		resolvedQuery, err := s.executionTree.workspace.ResolveQueryFromQueryProvider(queryProvider, runtimeArgs)
		if err != nil {
			return err
		}
		s.RawSQL = resolvedQuery.RawSQL
		s.executeSQL = resolvedQuery.ExecuteSQL
		s.Args = resolvedQuery.Args
	} else {
		// otherwise just resolve the args

		// merge the base args with the runtime args
		runtimeArgs, err = modconfig.MergeArgs(queryProvider, runtimeArgs)
		if err != nil {
			return err
		}

		args, err := modconfig.ResolveArgs(queryProvider, runtimeArgs)
		if err != nil {
			return err
		}
		s.Args = args
	}
	return nil
}

func (s *RuntimeDependencySubscriberImpl) populateParamDefaults(provider modconfig.QueryProvider) {
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
func (s *RuntimeDependencySubscriberImpl) buildRuntimeDependencyArgs() (*modconfig.QueryArgs, error) {
	res := modconfig.NewQueryArgs()

	log.Printf("[TRACE] %s: buildRuntimeDependencyArgs - %d runtime dependencies", s.resource.Name(), len(s.runtimeDependencies))

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

// populate the list of runtime dependencies that this run depends on
func (s *RuntimeDependencySubscriberImpl) setRuntimeDependencies() {
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

func (s *RuntimeDependencySubscriberImpl) hasRuntimeDependencies() bool {
	return len(s.runtimeDependencies)+len(s.baseRuntimeDependencies()) > 0
}

func (s *RuntimeDependencySubscriberImpl) baseRuntimeDependencies() map[string]*dashboardtypes.ResolvedRuntimeDependency {
	if s.baseDependencySubscriber == nil {
		return map[string]*dashboardtypes.ResolvedRuntimeDependency{}
	}
	return s.baseDependencySubscriber.runtimeDependencies
}

// override DashboardParentImpl.executeChildrenAsync to also execute 'withs' of our baseRun
func (s *RuntimeDependencySubscriberImpl) executeChildrenAsync(ctx context.Context) {
	// if we have a baseDependencySubscriber, execute it
	if s.baseDependencySubscriber != nil {
		go s.baseDependencySubscriber.executeWithsAsync(ctx)
	}

	// if this leaf run has children (including with runs) execute them asynchronously

	// set RuntimeDependenciesOnly if needed
	s.DashboardParentImpl.executeChildrenAsync(ctx)
}

// called when the args are resolved - if anyone is subscribing to the args value, publish
func (s *RuntimeDependencySubscriberImpl) argsResolved(args []any) {
	if s.baseDependencySubscriber != nil {
		s.baseDependencySubscriber.argsResolved(args)
	}
	s.runtimeDependencyPublisherImpl.argsResolved(args)
}
