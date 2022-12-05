package dashboardexecute

import (
	"context"
	"encoding/json"
	"fmt"
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
	Name             string                            `json:"name"`
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
	TimingResult     *queryresult.TimingResult         `json:"-"`
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
	name := getUniqueRunName(resource, executionTree)

	r := &LeafRun{
		RuntimeDependencyPublisherBase: *NewRuntimeDependencyPublisherBase(parent),
		Name:                           name,
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

	parsedName, err := modconfig.ParseResourceName(resource.Name())
	if err != nil {
		return nil, err
	}
	r.NodeType = parsedName.ItemType

	if err := r.addRuntimeDependencies(resource); err != nil {
		return nil, err
	}
	// if the node has no runtime dependencies, resolve the sql
	if len(r.runtimeDependencies) == 0 {
		if err := r.resolveSQLAndArgs(); err != nil {
			return nil, err
		}
	}
	// add r into execution tree
	executionTree.runs[r.Name] = r

	// if we have with blocks, create runs for them
	// BEFORE creating child runs
	withBlocks := r.DashboardNode.(modconfig.QueryProvider).GetWiths()
	if len(withBlocks) > 0 {
		// create the child runs
		err := r.createWithRuns(withBlocks, executionTree)
		if err != nil {
			return nil, err
		}
	}

	// if we have children (nodes/edges), create runs for them
	children := resource.GetChildren()
	if len(children) > 0 {
		// create the child runs
		err = r.createChildRuns(children, executionTree)
		if err != nil {
			return nil, err
		}
	}

	// create buffered child complete chan
	if childCount := len(children) + len(withBlocks); childCount > 0 {
		r.childComplete = make(chan dashboardtypes.DashboardNodeRun, childCount)
	}

	return r, nil
}

// resources (such as nodes/edges) may be reused by different parents - so wee need to give their LeafRuns unique names
func getUniqueRunName(resource modconfig.DashboardLeafNode, executionTree *DashboardExecutionTree) string {
	name := resource.Name()
	// check for uniqueness
	idx := 0
	for _, nameExists := executionTree.runs[name]; nameExists; idx++ {
		name = fmt.Sprintf("%s.%d", resource.Name(), idx)
	}
	return name
}

func (r *LeafRun) createChildRuns(children []modconfig.ModTreeItem, executionTree *DashboardExecutionTree) error {
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

	// start any `with` blocks
	r.executeWithRuns(ctx, r.childComplete)

	// we can either have children (i.e. edges/nodes) or we have sql/query
	// we have already validated that both are not set so no need to check here
	if len(r.children) > 0 {
		r.executeChildren(ctx)
	} else {
		if len(r.runtimeDependencies) > 0 {
			// if there are any unresolved runtime dependencies, wait for them
			if err := r.waitForRuntimeDependencies(ctx); err != nil {
				r.SetError(ctx, err)
				return
			}

			// ok now we have runtime dependencies, we can resolve the query
			if err := r.resolveSQLAndArgs(); err != nil {
				r.SetError(ctx, err)
				return
			}
		}
		if r.executeSQL == "" {
			r.SetError(ctx, fmt.Errorf("%s does not define query, SQL or nodes/edges", r.DashboardNode.Name()))
			return
		}
		r.executeQuery(ctx)
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

func (r *LeafRun) MarshalJSON() ([]byte, error) {

	// special case handling for NodeAndEdgeProvider
	_, isNodeAndEdgeProvider := r.DashboardNode.(modconfig.NodeAndEdgeProvider)
	if isNodeAndEdgeProvider {
		return r.marshalNodeAndEdgeProvider()
	}

	// just marshal as normal
	type Alias LeafRun
	return json.Marshal(struct{ *Alias }{(*Alias)(r)})
}

// we need custom JSON serialisation for NodeAndEdgeProviders
// This is because the name of the nodes and edges, which appears under properties,
// must be populated with the names of the node and edge LeafRuns, rather than the nodes and edge resources.
// These may be the same, but as the nodes/edges may be reused we ensure the run names are unique
// The panels in the panel map will be keyed by run-name - so it is vital that the nodes and edges lists
// correspond to the panel keys.
func (r *LeafRun) marshalNodeAndEdgeProvider() ([]byte, error) {
	type Alias LeafRun
	// embed the run in a struct, wiuth an additional 'Properties' property.
	// This will overwrite the `properties` value serialized from the underlying run
	s := &struct {
		Properties map[string]any `json:"properties"`
		*Alias
	}{
		Alias:      (*Alias)(r),
		Properties: make(map[string]any),
	}

	// add the node/edge child runs into the properties map, under the keys 'nodes'/'edges'
	for _, c := range r.GetChildren() {
		childResource := c.(*LeafRun).DashboardNode
		var childKey string

		switch childResource.(type) {
		case *modconfig.DashboardNode:
			childKey = "nodes"
		case *modconfig.DashboardEdge:
			childKey = "edges"
		}
		// add this child to the appropriate array
		target, _ := s.Properties[childKey].([]string)
		if target == nil {
			target = []string{}
		}
		s.Properties[childKey] = append(target, c.GetName())
	}

	// now marshal/ the DashboardNode resource then unmarshal back into the properties map
	resourceJson, err := json.Marshal(r.DashboardNode)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resourceJson, &s.Properties); err != nil {
		return nil, err
	}

	// now marshal our modified struct
	return json.Marshal(s)
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
	// does this leaf run have any SQL to execute?
	if !queryProvider.RequiresExecution(queryProvider) {
		log.Printf("[TRACE] LeafRun '%s'does NOT require execution - returning", r.DashboardNode.Name())
		return nil
	}
	err := queryProvider.VerifyQuery(queryProvider)
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' VerifyQuery failed: %s", r.DashboardNode.Name(), err.Error())
		return err
	}

	// convert runtime dependencies into arg map
	runtimeArgs, err := r.buildRuntimeDependencyArgs()
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs failed: %s", r.DashboardNode.Name(), err.Error())
		return err
	}

	log.Printf("[TRACE] LeafRun '%s' built runtime args: %v", r.DashboardNode.Name(), runtimeArgs)

	resolvedQuery, err := r.executionTree.workspace.ResolveQueryFromQueryProvider(queryProvider, runtimeArgs)
	if err != nil {
		return err
	}
	r.RawSQL = resolvedQuery.RawSQL
	r.executeSQL = resolvedQuery.ExecuteSQL
	r.Args = resolvedQuery.Args
	return nil
}

// convert runtime dependencies into arg map
func (r *LeafRun) buildRuntimeDependencyArgs() (*modconfig.QueryArgs, error) {
	res := modconfig.NewQueryArgs()

	log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs - %d runtime dependencies", r.DashboardNode.Name(), len(r.runtimeDependencies))

	// if the runtime dependencies use position args, get the max index and ensure the args array is large enough
	maxArgIndex := -1
	for _, dep := range r.runtimeDependencies {
		if dep.Dependency.ArgIndex != nil && *dep.Dependency.ArgIndex > maxArgIndex {
			maxArgIndex = *dep.Dependency.ArgIndex
		}
	}
	if maxArgIndex != -1 {
		res.ArgList = make([]*string, maxArgIndex+1)
	}

	// build map of default params
	for _, dep := range r.runtimeDependencies {
		if dep.Dependency.ArgName != nil {
			err := res.SetNamedArgVal(dep.Value, *dep.Dependency.ArgName)
			if err != nil {
				return nil, err
			}

		} else {
			if dep.Dependency.ArgIndex == nil {
				return nil, fmt.Errorf("invalid runtime dependency - both ArgName and ArgIndex are nil ")
			}
			err := res.SetPositionalArgVal(dep.Value, *dep.Dependency.ArgIndex)
			if err != nil {
				return nil, err
			}
		}
	}
	return res, nil
}

// if this leaf run has a query or sql, execute it now
func (r *LeafRun) executeQuery(ctx context.Context) {
	log.Printf("[TRACE] LeafRun '%s' SQL resolved, executing", r.DashboardNode.Name())

	queryResult, err := r.executionTree.client.ExecuteSync(ctx, r.executeSQL, r.Args...)
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' query failed: %s", r.DashboardNode.Name(), err.Error())
		// set the error status on the counter - this will raise counter error event
		r.SetError(ctx, err)
		return

	}
	log.Printf("[TRACE] LeafRun '%s' complete", r.DashboardNode.Name())

	r.Data = dashboardtypes.NewLeafData(queryResult)
	r.TimingResult = queryResult.TimingResult
	// set complete status on counter - this will raise counter complete event
	r.SetComplete(ctx)
}

// if this leaf run has children (nodes/edges), execute them
func (r *LeafRun) executeChildren(ctx context.Context) {
	for _, c := range r.children {
		go c.Execute(ctx)
	}
	// wait for children to complete
	var errors []error

	for !r.ChildrenComplete() {
		log.Printf("[TRACE] run %s waiting for children", r.Name)
		completeChild := <-r.childComplete
		log.Printf("[TRACE] run %s got child complete", r.Name)
		if completeChild.GetRunStatus() == dashboardtypes.DashboardRunError {
			errors = append(errors, completeChild.GetError())
		}
		// fall through to recheck ChildrenComplete
	}

	log.Printf("[TRACE] run %s ALL children complete", r.Name)
	// so all children have completed - check for errors
	// TODO format better error
	err := error_helpers.CombineErrors(errors...)
	// combine child data even if there is an error
	r.combineChildData()
	if err == nil {
		// set complete status on dashboard
		r.SetComplete(ctx)
	} else {
		r.SetError(ctx, err)
	}
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
		if p.Name == paramName {
			return true
		}
	}
	return false
}

//func (r *LeafRun) subscribeToParamValue(name string, path *modconfig.ParsedPropertyPath) chan *dashboardtypes.ResolvedRuntimeDependencyValue {
//
//}
//
//func (r *LeafRun) subscribeToWithValue(name string, path *modconfig.ParsedPropertyPath) chan *dashboardtypes.ResolvedRuntimeDependencyValue {
//	log.Printf("[TRACE] subscribeToWithValue %s", name)
//	// make a channel (buffer to avoid potential sync issues)
//	valueChannel := make(chan *dashboardtypes.ResolvedRuntimeDependencyValue, 1)
//
//	// subscribe, passing a function which invokes getWithValue to resolve the required with value
//	r.withValueSubscriptions[name] = append(r.withValueSubscriptions[name], func(result *dashboardtypes.WithResult) {
//		// the WithResult includes an error field indicating an error running the with query
//		resolvedResult := &dashboardtypes.ResolvedRuntimeDependencyValue{Error: result.Error}
//		// if there was NO error, resolve the required result value
//		if result.Error == nil {
//			resolvedResult.Value, resolvedResult.Error = r.getWithValue(name, result, path)
//		}
//		valueChannel <- resolvedResult
//		close(valueChannel)
//	})
//	return valueChannel
//
//}
//
//func (r *LeafRun) publishWithValues(name string, result *dashboardtypes.WithResult) {
//	for _, f := range r.withValueSubscriptions[name] {
//		f(result)
//	}
//	// clear subscriptions
//	delete(r.withValueSubscriptions, name)
//}
