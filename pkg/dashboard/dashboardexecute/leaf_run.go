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
	RuntimeDependencyPublisherBase
	Title            string                            `json:"title,omitempty"`
	Width            int                               `json:"width,omitempty"`
	Type             string                            `cty:"type" hcl:"type" column:"type,text" json:"display_type,omitempty"`
	Display          string                            `cty:"display" hcl:"display" json:"display,omitempty"`
	RawSQL           string                            `json:"sql,omitempty"`
	Data             *dashboardtypes.LeafData          `json:"data,omitempty"`
	ErrorString      string                            `json:"error,omitempty"`
	DashboardNode    modconfig.DashboardLeafNode       `json:"properties,omitempty"`
	NodeType         string                            `json:"panel_type"`
	Status           dashboardtypes.DashboardRunStatus `json:"status"`
	DashboardName    string                            `json:"dashboard"`
	SourceDefinition string                            `json:"source_definition"`
	// a list of the (scoped) names of any `withs` that we rely on
	DependencyWiths []string                  `json:"withs,omitempty"`
	TimingResult    *queryresult.TimingResult `json:"-"`
	// child runs (nodes/edges)
	children      []dashboardtypes.DashboardNodeRun
	executeSQL    string
	error         error
	parent        dashboardtypes.DashboardNodeParent
	executionTree *DashboardExecutionTree
	childComplete chan dashboardtypes.DashboardNodeRun
}

func (r *LeafRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	return &dashboardtypes.SnapshotTreeNode{
		Name:     r.Name,
		NodeType: r.NodeType,
	}
}

func NewLeafRun(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardNodeParent, executionTree *DashboardExecutionTree) (*LeafRun, error) {
	name := resource.Name()

	r := &LeafRun{
		RuntimeDependencyPublisherBase: *NewRuntimeDependencyPublisherBase(name, parent),
		Title:                          resource.GetTitle(),
		Width:                          resource.GetWidth(),
		Type:                           resource.GetType(),
		Display:                        resource.GetDisplay(),
		DashboardNode:                  resource,
		DashboardName:                  executionTree.dashboardName,

		SourceDefinition: resource.GetMetadata().SourceDefinition,
		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		Status:        dashboardtypes.DashboardRunComplete,
		executionTree: executionTree,
		parent:        parent,
	}
	parsedName, err := modconfig.ParseResourceName(name)
	if err != nil {
		return nil, err
	}
	r.NodeType = parsedName.ItemType

	// is the resource  a query provider
	if queryProvider, ok := r.DashboardNode.(modconfig.QueryProvider); ok {
		// set params
		r.Params = queryProvider.GetParams()
		// set Status
		// if the query provider resource requires execution OR we have children, set status to "ready"
		// (indicating we must be executed)
		if queryProvider.RequiresExecution(queryProvider) || len(resource.GetChildren()) > 0 {
			r.Status = dashboardtypes.DashboardRunReady
		}
	}

	// if the resource is a runtime dependency provider, create with runs if needed and set runtime dependencies
	var withBlocks []*modconfig.DashboardWith
	if rdp, ok := r.DashboardNode.(modconfig.RuntimeDependencyProvider); ok {
		// if we have with blocks, create runs for them
		// BEFORE creating child runs, and before adding runtime dependencies
		withBlocks = rdp.GetWiths()
		if len(withBlocks) > 0 {
			// create the child runs
			err := r.createWithRuns(withBlocks, executionTree)
			if err != nil {
				return nil, err
			}
		}

		if err := r.resolveRuntimeDependencies(rdp); err != nil {
			return nil, err
		}
	}

	// if the node has no runtime dependencies, resolve the sql
	if len(r.runtimeDependencies) == 0 {
		if err := r.resolveSQLAndArgs(); err != nil {
			return nil, err
		}
	}
	// add r into execution tree
	executionTree.runs[r.Name] = r

	// if we have children (nodes/edges), create runs for them
	children := resource.GetChildren()
	// create the child runs
	err = r.createChildRuns(children, executionTree)
	if err != nil {
		return nil, err
	}

	// create buffered child complete chan
	if childCount := len(children) + len(withBlocks); childCount > 0 {
		r.childComplete = make(chan dashboardtypes.DashboardNodeRun, childCount)
	}

	return r, nil
}

func (r *LeafRun) createChildRuns(children []modconfig.ModTreeItem, executionTree *DashboardExecutionTree) error {
	if len(children) == 0 {
		return nil
	}

	r.children = make([]dashboardtypes.DashboardNodeRun, len(children))
	var errors []error

	// if the leaf run has children (nodes/edges) create a run for this too
	for i, c := range children {
		childRun, err := NewLeafRun(c.(modconfig.DashboardLeafNode), r, executionTree)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		r.children[i] = childRun
	}
	return error_helpers.CombineErrors(errors...)
}

func (r *LeafRun) createWithRuns(withs []*modconfig.DashboardWith, executionTree *DashboardExecutionTree) error {
	for _, w := range withs {
		withRun, err := NewLeafRun(w, r, executionTree)
		if err != nil {
			return err
		}
		r.withRuns[w.UnqualifiedName] = withRun
	}
	return nil
}

// Initialise implements DashboardRunNode
func (r *LeafRun) Initialise(ctx context.Context) {}

// Execute implements DashboardRunNode
func (r *LeafRun) Execute(ctx context.Context) {
	// if there is nothing to do, return
	if r.Status == dashboardtypes.DashboardRunComplete {
		return
	}

	log.Printf("[TRACE] LeafRun '%s' Execute()", r.DashboardNode.Name())

	// to get here, we must be a query provider

	// TODO validate we have either children or query
	//r.SetError(ctx, fmt.Errorf("%s does not define query, SQL or nodes/edges", r.DashboardNode.Name()))

	// if we have children and with runs, start them asyncronously (they may block waiting for our runtime dependencies)
	r.executeChildrenAsync(ctx)

	// now collect data from children and withs
	doneChan := r.collectChildDataAsync(ctx)

	// now wait for any runtime dependencies then resolve args and params
	// (it is possible to have params but no sql)
	if len(r.runtimeDependencies) > 0 {
		// if there are any unresolved runtime dependencies, wait for them
		if err := r.waitForRuntimeDependencies(ctx); err != nil {
			r.SetError(ctx, err)
			return
		}

		// populate the names of any withs we depend on
		r.setDependencyWiths()
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
		// set complete status on dashboard
		r.SetComplete(ctx)
	} else {

		r.SetError(ctx, err)
	}
}

// GetName implements DashboardNodeRun
func (r *LeafRun) GetName() string {
	return r.Name
}

// GetParent implements DashboardNodeRun
func (r *LeafRun) GetParent() dashboardtypes.DashboardNodeParent {
	return r.parent
}

// GetRunStatus implements DashboardNodeRun
func (r *LeafRun) GetRunStatus() dashboardtypes.DashboardRunStatus {
	return r.Status
}

// SetError implements DashboardNodeRun
func (r *LeafRun) SetError(ctx context.Context, err error) {
	r.error = err
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
	})
	r.parent.ChildCompleteChan() <- r
}

// GetError implements DashboardNodeRun
func (r *LeafRun) GetError() error {
	return r.error
}

// SetComplete implements DashboardNodeRun
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

// RunComplete implements DashboardNodeRun
func (r *LeafRun) RunComplete() bool {
	return r.Status == dashboardtypes.DashboardRunComplete || r.Status == dashboardtypes.DashboardRunError
}

// GetChildren implements DashboardNodeRun
func (r *LeafRun) GetChildren() []dashboardtypes.DashboardNodeRun {
	return r.children
}

// ChildrenComplete implements DashboardNodeRun
func (r *LeafRun) ChildrenComplete() bool {
	for _, child := range r.children {
		if !child.RunComplete() {
			return false
		}
	}
	return true
}

// IsSnapshotPanel implements SnapshotPanel
func (*LeafRun) IsSnapshotPanel() {}

// GetTitle implements DashboardNodeRun
func (r *LeafRun) GetTitle() string {
	return r.Title
}

// GetInputsDependingOn implements DashboardNodeRun
// return nothing for LeafRun
func (r *LeafRun) GetInputsDependingOn(changedInputName string) []string { return nil }

// ChildCompleteChan implements DashboardNodeParent
func (r *LeafRun) ChildCompleteChan() chan dashboardtypes.DashboardNodeRun {
	return r.childComplete
}

func (r *LeafRun) waitForRuntimeDependencies(ctx context.Context) error {
	log.Printf("[TRACE] LeafRun '%s' waitForRuntimeDependencies", r.DashboardNode.Name())
	for _, resolvedDependency := range r.runtimeDependencies {
		// check whether the dependency is available
		err := resolvedDependency.Resolve()
		if err != nil {
			return err
		}
	}

	if len(r.runtimeDependencies) > 0 {
		log.Printf("[TRACE] LeafRun '%s' all runtime dependencies ready", r.DashboardNode.Name())
	}
	return nil
}

// resolve the sql for this leaf run into the source sql (i.e. NOT the prepared statement name) and resolved args
func (r *LeafRun) resolveSQLAndArgs() error {
	log.Printf("[TRACE] LeafRun '%s' resolveSQLAndArgs", r.DashboardNode.Name())
	queryProvider, ok := r.DashboardNode.(modconfig.QueryProvider)
	if !ok {
		// not a query provider - nothing to do
		return nil
	}

	err := queryProvider.VerifyQuery(queryProvider)
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' VerifyQuery failed: %s", r.DashboardNode.Name(), err.Error())
		return err
	}

	// convert arg runtime dependencies into arg map
	runtimeArgs, err := r.buildRuntimeDependencyArgs()
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs failed: %s", r.DashboardNode.Name(), err.Error())
		return err
	}

	// now if any param defaults had runtime depdencies, populate them
	r.populateParamDefaults(queryProvider)

	log.Printf("[TRACE] LeafRun '%s' built runtime args: %v", r.DashboardNode.Name(), runtimeArgs)

	// does this leaf run have any SQL to execute?
	// TODO split this into resolve query and resolve args - we may have args but no query
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

	log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs - %d runtime dependencies", r.DashboardNode.Name(), len(r.runtimeDependencies))

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

// if this leaf run has a query or sql, execute it now
func (r *LeafRun) executeQuery(ctx context.Context) error {
	log.Printf("[TRACE] LeafRun '%s' SQL resolved, executing", r.DashboardNode.Name())

	queryResult, err := r.executionTree.client.ExecuteSync(ctx, r.executeSQL, r.Args...)
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' query failed: %s", r.DashboardNode.Name(), err.Error())
		return err

	}
	log.Printf("[TRACE] LeafRun '%s' complete", r.DashboardNode.Name())

	r.Data = dashboardtypes.NewLeafData(queryResult)
	r.TimingResult = queryResult.TimingResult
	return nil
}

// if this leaf run has children (nodes/edges), and with runs execute them asynchronously
func (r *LeafRun) executeChildrenAsync(ctx context.Context) {
	for _, c := range r.children {
		go c.Execute(ctx)
	}

	for _, w := range r.withRuns {
		go w.Execute(ctx)
	}
}
func (r *LeafRun) collectChildDataAsync(ctx context.Context) chan error {
	var doneChan = make(chan error)
	go func() {
		// wait for children to complete
		var errors []error

		for !r.ChildrenComplete() {
			log.Printf("[TRACE] run %s waiting for children", r.Name)
			completeChild := <-r.childComplete
			log.Printf("[TRACE] run %s got child complete", r.Name)
			if completeChild.GetRunStatus() == dashboardtypes.DashboardRunError {
				errors = append(errors, completeChild.GetError())
			} else {
				// if this is a with, set with data
				if leafRun, ok := completeChild.(*LeafRun); ok && leafRun.NodeType == modconfig.BlockTypeWith {
					r.setWithValue(leafRun)
				}
			}
			// fall through to recheck ChildrenComplete
		}

		log.Printf("[TRACE] run %s ALL children complete", r.Name)
		// so all children have completed - check for errors
		// TODO format better error
		err := error_helpers.CombineErrors(errors...)
		// combine child data even if there is an error
		r.combineChildData()
		doneChan <- err
	}()
	return doneChan
}

func (r *LeafRun) combineChildData() {
	r.Data = &dashboardtypes.LeafData{}
	// build map of columns for the schema
	schemaMap := make(map[string]*queryresult.ColumnDef)
	for _, c := range r.children {
		childLeafRun := c.(*LeafRun)
		data := childLeafRun.Data
		if data == nil {
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

func (r *LeafRun) hasParam(paramName string) bool {
	for _, p := range r.Params {
		if p.ShortName == paramName {
			return true
		}
	}
	return false
}

// populate the list of `withs` that this run depends on
func (r *LeafRun) setDependencyWiths() {
	for _, d := range r.runtimeDependencies {
		if d.Dependency.PropertyPath.ItemType == modconfig.BlockTypeWith {
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

