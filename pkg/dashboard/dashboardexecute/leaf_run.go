package dashboardexecute

import (
	"context"
	"fmt"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"golang.org/x/exp/maps"
	"log"
)

// LeafRun is a struct representing the execution of a leaf dashboard node
type LeafRun struct {
	RuntimeDependencyPublisherImpl

	RawSQL   string                      `json:"sql,omitempty"`
	Data     *dashboardtypes.LeafData    `json:"data,omitempty"`
	Resource modconfig.DashboardLeafNode `json:"properties,omitempty"`
	// a list of the (scoped) names of any `withs` that we rely on
	DependencyWiths []string                  `json:"withs,omitempty"`
	TimingResult    *queryresult.TimingResult `json:"-"`
	executeSQL      string
	onComplete      func()
}

func (r *LeafRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	return &dashboardtypes.SnapshotTreeNode{
		Name:     r.Name,
		NodeType: r.NodeType,
	}
}

func NewLeafRun(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) (*LeafRun, error) {
	r := &LeafRun{
		// create RuntimeDependencyPublisherImpl- this handles 'with' run creation and resolving runtime dependency resolution
		RuntimeDependencyPublisherImpl: *NewRuntimeDependencyPublisherImpl(resource, parent, executionTree),
		Resource:                       resource,
	}
	err := r.initRuntimeDependencies()
	if err != nil {
		return nil, err
	}

	r.NodeType = resource.BlockType()

	// if the node has no runtime dependencies, resolve the sql
	if len(r.runtimeDependencies) == 0 {
		if err := r.resolveSQLAndArgs(); err != nil {
			return nil, err
		}
	}
	// add r into execution tree
	executionTree.runs[r.Name] = r

	// if we have children (nodes/edges), create runs for them
	err = r.createChildRuns(executionTree)
	if err != nil {
		return nil, err
	}

	// create buffered channel for children to report their completion
	r.createChildCompleteChan()

	// populate the names of any withs we depend on
	r.setDependencyWiths()

	return r, nil
}

func (r *LeafRun) createChildRuns(executionTree *DashboardExecutionTree) error {
	children := r.Resource.GetChildren()
	if len(children) == 0 {
		return nil
	}

	r.children = make([]dashboardtypes.DashboardTreeRun, len(children))
	var errors []error

	// if the leaf run has children (nodes/edges) create a run for this too
	for i, c := range children {
		// TODO [node_reuse] what about with nodes - only relevant when running base withs
		childRun, err := NewLeafRun(c.(modconfig.DashboardLeafNode), r, executionTree)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		r.children[i] = childRun
	}
	return error_helpers.CombineErrors(errors...)
}

// Execute implements DashboardRunNode
func (r *LeafRun) Execute(ctx context.Context) {
	defer func() {
		// call our oncomplete is we have one
		// (this is used to collect 'with' data and propagate errors)
		if r.onComplete != nil {
			r.onComplete()
		}
	}()

	// if there is nothing to do, return
	if r.Status == dashboardtypes.DashboardRunComplete {
		return
	}

	log.Printf("[TRACE] LeafRun '%s' Execute()", r.Resource.Name())

	// to get here, we must be a query provider

	// if we have children and with runs, start them asyncronously (they may block waiting for our runtime dependencies)
	r.executeChildrenAsync(ctx)

	// start a goroutine to wait for children to complete
	doneChan := r.waitForChildren()

	// now wait for any runtime dependencies then resolve args and params
	// (it is possible to have params but no sql)
	if len(r.runtimeDependencies) > 0 {
		// if there are any unresolved runtime dependencies, wait for them
		if err := r.waitForRuntimeDependencies(); err != nil {
			r.SetError(ctx, err)
			return
		}

		// ok now we have runtime dependencies, we can resolve the query
		if err := r.resolveSQLAndArgs(); err != nil {
			r.SetError(ctx, err)
			return
		}
	}

	// if we have sql to execute, do it now
	if r.executeSQL != "" {
		if err := r.executeQuery(ctx); err != nil {
			r.SetError(ctx, err)
			return
		}

	}

	// wait for all children and withs
	err := <-doneChan
	if err == nil {
		// aggregate our child data
		r.combineChildData()
		// set complete status on dashboard
		r.SetComplete(ctx)
	} else {

		r.SetError(ctx, err)
	}
}

// SetError implements DashboardTreeRun
func (r *LeafRun) SetError(ctx context.Context, err error) {
	log.Printf("[TRACE] %s SetError err %v", r.Name, err)
	r.err = err
	// error type does not serialise to JSON so copy into a string
	r.ErrorString = err.Error()
	r.Status = dashboardtypes.DashboardRunError
	// increment error count for snapshot hook
	statushooks.SnapshotError(ctx)
	// raise counter error event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeError{
		LeafNode:    r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
		Error:       err,
	})
	r.parent.ChildCompleteChan() <- r
}

// SetComplete implements DashboardTreeRun
func (r *LeafRun) SetComplete(ctx context.Context) {
	r.Status = dashboardtypes.DashboardRunComplete
	// raise counter complete event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeComplete{
		LeafNode:    r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
	})

	// call snapshot hooks with progress
	statushooks.UpdateSnapshotProgress(ctx, 1)

	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// IsSnapshotPanel implements SnapshotPanel
func (*LeafRun) IsSnapshotPanel() {}

func (r *LeafRun) waitForRuntimeDependencies() error {
	log.Printf("[TRACE] LeafRun '%s' waitForRuntimeDependencies", r.Resource.Name())
	for _, resolvedDependency := range r.runtimeDependencies {
		// check whether the dependency is available
		err := resolvedDependency.Resolve()
		if err != nil {
			return err
		}
	}

	if len(r.runtimeDependencies) > 0 {
		log.Printf("[TRACE] LeafRun '%s' all runtime dependencies ready", r.Resource.Name())
	}
	return nil
}

// resolve the sql for this leaf run into the source sql (i.e. NOT the prepared statement name) and resolved args
func (r *LeafRun) resolveSQLAndArgs() error {
	log.Printf("[TRACE] LeafRun '%s' resolveSQLAndArgs", r.Resource.Name())
	queryProvider, ok := r.Resource.(modconfig.QueryProvider)
	if !ok {
		// not a query provider - nothing to do
		return nil
	}

	// convert arg runtime dependencies into arg map
	runtimeArgs, err := r.buildRuntimeDependencyArgs()
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs failed: %s", r.Resource.Name(), err.Error())
		return err
	}

	// now if any param defaults had runtime depdencies, populate them
	r.populateParamDefaults(queryProvider)

	log.Printf("[TRACE] LeafRun '%s' built runtime args: %v", r.Resource.Name(), runtimeArgs)

	// does this leaf run have any SQL to execute?
	// TODO [node_reuse] split this into resolve query and resolve args - we may have args but no query
	if queryProvider.RequiresExecution(queryProvider) {
		resolvedQuery, err := r.executionTree.workspace.ResolveQueryFromQueryProvider(queryProvider, runtimeArgs)
		if err != nil {
			return err
		}
		r.RawSQL = resolvedQuery.RawSQL
		r.executeSQL = resolvedQuery.ExecuteSQL
		r.Args = resolvedQuery.Args
	}
	//}
	return nil
}

// convert runtime dependencies into arg map
func (r *LeafRun) buildRuntimeDependencyArgs() (*modconfig.QueryArgs, error) {
	res := modconfig.NewQueryArgs()

	log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs - %d runtime dependencies", r.Resource.Name(), len(r.runtimeDependencies))

	// if the runtime dependencies use position args, get the max index and ensure the args array is large enough
	maxArgIndex := -1
	// build list of all args runtime dependencies
	argRuntimeDependencies := r.FindRuntimeDependenciesForParent(modconfig.AttributeArgs)

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

func (r *LeafRun) hasParam(paramName string) bool {
	for _, p := range r.Params {
		if p.ShortName == paramName {
			return true
		}
	}
	return false
}

// if this leaf run has a query or sql, execute it now
func (r *LeafRun) executeQuery(ctx context.Context) error {
	log.Printf("[TRACE] LeafRun '%s' SQL resolved, executing", r.Resource.Name())

	queryResult, err := r.executionTree.client.ExecuteSync(ctx, r.executeSQL, r.Args...)
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' query failed: %s", r.Resource.Name(), err.Error())
		return err

	}
	log.Printf("[TRACE] LeafRun '%s' complete", r.Resource.Name())

	r.Data = dashboardtypes.NewLeafData(queryResult)
	r.TimingResult = queryResult.TimingResult
	return nil
}

func (r *LeafRun) combineChildData() {
	// we either have children OR a query
	// if there are no children, do nothing
	if len(r.children) == 0 {
		return
	}
	// create empty data to populate
	r.Data = &dashboardtypes.LeafData{}
	// build map of columns for the schema
	schemaMap := make(map[string]*queryresult.ColumnDef)
	for _, c := range r.children {
		childLeafRun := c.(*LeafRun)
		data := childLeafRun.Data
		// if there is no data or this is a 'with', skip
		if data == nil || childLeafRun.Resource.BlockType() == modconfig.BlockTypeWith {
			continue
		}
		for _, s := range data.Columns {
			if _, ok := schemaMap[s.Name]; !ok {
				schemaMap[s.Name] = s
			}
		}
		r.Data.Rows = append(r.Data.Rows, data.Rows...)
	}
	r.Data.Columns = maps.Values(schemaMap)
}

// populate the list of `withs` that this run depends on
func (r *LeafRun) setDependencyWiths() {
	for _, d := range r.runtimeDependencies {
		if d.Dependency.PropertyPath.ItemType == modconfig.BlockTypeWith {
			// add to DependencyWiths using ScopedName, i.e. <parent FullName>.<with UnqualifiedName>.
			// we do this as there may be a with from a base resource with a clashing with name
			// NOTE: this must be consistent with the naming in RuntimeDependencyPublisherImpl.createWithRuns
			r.DependencyWiths = append(r.DependencyWiths, d.ScopedName())
		}
	}
}

func (r *LeafRun) populateParamDefaults(provider modconfig.QueryProvider) {
	paramDefs := provider.GetParams()
	for _, paramDef := range paramDefs {
		if dep := r.FindRuntimeDependencyForParent(paramDef.UnqualifiedName); dep != nil {
			// assuming the default property is the target, set the default
			if typehelpers.SafeString(dep.Dependency.TargetPropertyName) == "default" {
				paramDef.SetDefault(dep.Value)
			}
		}
	}
}
