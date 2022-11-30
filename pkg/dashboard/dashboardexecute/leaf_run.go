package dashboardexecute

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
	"strconv"
	"sync"

	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"golang.org/x/exp/maps"
)

// LeafRun is a struct representing the execution of a leaf dashboard node
type LeafRun struct {
	Name             string                            `json:"name"`
	Title            string                            `json:"title,omitempty"`
	Width            int                               `json:"width,omitempty"`
	Type             string                            `cty:"type" hcl:"type" column:"type,text" json:"display_type,omitempty"`
	Display          string                            `cty:"display" hcl:"display" json:"display,omitempty"`
	RawSQL           string                            `json:"sql,omitempty"`
	Args             []any                             `json:"args,omitempty"`
	Params           []*modconfig.ParamDef             `json:"params,omitempty"`
	Data             *dashboardtypes.LeafData          `json:"data,omitempty"`
	ErrorString      string                            `json:"error,omitempty"`
	DashboardNode    modconfig.DashboardLeafNode       `json:"properties,omitempty"`
	NodeType         string                            `json:"panel_type"`
	Status           dashboardtypes.DashboardRunStatus `json:"status"`
	DashboardName    string                            `json:"dashboard"`
	SourceDefinition string                            `json:"source_definition"`
	TimingResult     *queryresult.TimingResult         `json:"-"`
	// child runs (nodes/edges)
	children            []dashboardtypes.DashboardNodeRun
	executeSQL          string
	error               error
	parent              dashboardtypes.DashboardNodeParent
	executionTree       *DashboardExecutionTree
	runtimeDependencies map[string]*ResolvedRuntimeDependency
	childComplete       chan dashboardtypes.DashboardNodeRun
	withValues          map[string]*dashboardtypes.LeafData
	withValueMutex      sync.Mutex
	withRuns            []*LeafRun
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
		Name:             name,
		Title:            resource.GetTitle(),
		Width:            resource.GetWidth(),
		Type:             resource.GetType(),
		Display:          resource.GetDisplay(),
		DashboardNode:    resource,
		DashboardName:    executionTree.dashboardName,
		SourceDefinition: resource.GetMetadata().SourceDefinition,
		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		Status:              dashboardtypes.DashboardRunComplete,
		executionTree:       executionTree,
		parent:              parent,
		runtimeDependencies: make(map[string]*ResolvedRuntimeDependency),

		withValues: make(map[string]*dashboardtypes.LeafData),
	}

	parsedName, err := modconfig.ParseResourceName(resource.Name())
	if err != nil {
		return nil, err
	}
	r.NodeType = parsedName.ItemType

	// determine whether we need to execute this node or its children
	r.setStatus()

	r.addRuntimeDependencies()
	// if the node has no runtime dependencies, resolve the sql
	if len(r.runtimeDependencies) == 0 {
		if err := r.resolveSQL(); err != nil {
			return nil, err
		}
	}
	// add r into execution tree
	executionTree.runs[r.Name] = r

	// if we have children (nodes/edges), create runs for them
	children := resource.GetChildren()
	if len(children) > 0 {
		// create the child runs
		err := r.createChildRuns(children, executionTree)
		if err != nil {
			return nil, err
		}
	}

	// if we have with blocks, create runs for them
	withBlocks := r.DashboardNode.(modconfig.QueryProvider).GetWiths()
	if len(withBlocks) > 0 {
		// create the child runs
		err := r.createWithRuns(withBlocks, executionTree)
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

// if this node has runtime dependencies, create runtime dependency instances which we use to resolve the values
func (r *LeafRun) addRuntimeDependencies() {
	// only QueryProvider resources support runtime dependencies
	queryProvider, ok := r.DashboardNode.(modconfig.QueryProvider)
	if !ok {
		return
	}
	runtimeDependencies := queryProvider.GetRuntimeDependencies()
	for n, d := range runtimeDependencies {
		// read name and dep into local loop vars to ensure correct value used when getValueFunc is invoked
		name := n
		dep := d
		// determine the function to use to retrieve the runtime dependency value
		var getValueFunc func(string) (any, error)
		switch dep.PropertyPath.ItemType {
		case modconfig.BlockTypeWith:
			getValueFunc = func(name string) (any, error) {
				return r.getWithValue(name, dep.PropertyPath)
			}
		case modconfig.BlockTypeInput:
			getValueFunc = r.executionTree.GetInputValue
		}
		r.runtimeDependencies[name] = NewResolvedRuntimeDependency(dep, getValueFunc)
	}
	// if the parent is a leaf run, we must be a node or an edge, inherit our parent runtime dependencies
	// NOTE: UNLESS we are a 'with' run
	if _, isWith := r.DashboardNode.(*modconfig.DashboardWith); !isWith {
		if parentLeafRun, ok := r.parent.(*LeafRun); ok {
			for name, dep := range parentLeafRun.runtimeDependencies {
				if _, ok := r.runtimeDependencies[name]; !ok {
					r.runtimeDependencies[name] = dep
				}
			}
		}
	}
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
	r.withRuns = make([]*LeafRun, len(withs))

	for i, w := range withs {
		withRun, err := NewLeafRun(w, r, executionTree)
		if err != nil {
			return err
		}
		r.withRuns[i] = withRun
	}
	return nil
}

// if we have a query provider which requires execution OR we have children, set status to ready
func (r *LeafRun) setStatus() {
	resource := r.DashboardNode
	if provider, ok := resource.(modconfig.QueryProvider); ok {
		if provider.RequiresExecution(provider) || len(resource.GetChildren()) > 0 {
			r.Status = dashboardtypes.DashboardRunReady
		}
	}
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

	// start all `with` blocks
	if len(r.withRuns) > 0 {
		r.executeWithRuns(ctx)
	}

	// TODO KAI handle error in with block

	// if there are any unresolved runtime dependencies, wait for them
	if len(r.runtimeDependencies) > 0 {
		if err := r.waitForRuntimeDependencies(ctx); err != nil {
			r.SetError(ctx, err)
			return
		}

		// ok now we have runtime dependencies, we can resolve the query
		if err := r.resolveSQL(); err != nil {
			r.SetError(ctx, err)
			return
		}
	}

	// we can either have children (i.e. edges/nodes) or we have sql/query
	// we have already validated that both are not set so no need to check here
	if len(r.children) > 0 {
		r.executeChildren(ctx)
	} else {
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

func (r *LeafRun) withComplete() bool {
	for _, w := range r.withRuns {
		if !w.RunComplete() {
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

	// special case handling for EdgeAndNodeProvider
	_, isEdgeAndNodeProvider := r.DashboardNode.(modconfig.EdgeAndNodeProvider)
	if isEdgeAndNodeProvider {
		return r.marshalEdgeAndNodeProvider()
	}

	// just marshal as normal
	type Alias LeafRun
	return json.Marshal(struct{ *Alias }{(*Alias)(r)})
}

// we need custom JSON serialisation for EdgeAndNodeProviders
// This is because the name of the nodes and edges, which appears under properties,
// must be populated with the names of the node and edge LeafRuns, rather than the nodes and edge resources.
// These may be the same, but as the nodes/edges may be reused we ensure the run names are unique
// The panels in the panel map will be keyed by run-name - so it is vital that the nodes and edges lists
// correspond to the panel keys.
func (r *LeafRun) marshalEdgeAndNodeProvider() ([]byte, error) {
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
		isResolved, err := resolvedDependency.Resolve()
		if err != nil {
			return err
		}

		if isResolved {
			// this one is available
			continue
		}
		log.Printf("[TRACE] waitForRuntimeDependency %s", resolvedDependency.dependency.String())
		if err := r.executionTree.waitForRuntimeDependency(ctx, resolvedDependency.dependency); err != nil {
			return err
		}

		log.Printf("[TRACE] dependency %s should be available", resolvedDependency.dependency.String())

		// now again resolve the dependency value - this sets the arg to have the runtime dependency value
		isResolved, err = resolvedDependency.Resolve()
		if err != nil {
			return err
		}
		if !isResolved {
			log.Printf("[TRACE] dependency %s not resolved after waitForRuntimeDependency returned", resolvedDependency.dependency.String())
			// should now be resolved`
			return fmt.Errorf("dependency %s not resolved after waitForRuntimeDependency returned", resolvedDependency.dependency.String())
		}
	}

	if len(r.runtimeDependencies) > 0 {
		log.Printf("[TRACE] LeafRun '%s' all runtime dependencies ready", r.DashboardNode.Name())
	}
	return nil
}

// resolve the sql for this leaf run into the source sql (i.e. NOT the prepared statement name) and resolved args
func (r *LeafRun) resolveSQL() error {
	log.Printf("[TRACE] LeafRun '%s' resolveSQL", r.DashboardNode.Name())
	queryProvider, ok := r.DashboardNode.(modconfig.QueryProvider)
	if !ok {
		// not a query provider - nothing to do
		return nil
	}

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
	r.Params = resolvedQuery.Params
	return nil
}

func (r *LeafRun) buildRuntimeDependencyArgs() (*modconfig.QueryArgs, error) {
	res := modconfig.NewQueryArgs()

	log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs - %d runtime dependencies", r.DashboardNode.Name(), len(r.runtimeDependencies))

	// if the runtime dependencies use position args, get the max index and ensure the args array is large enough
	maxArgIndex := -1
	for _, dep := range r.runtimeDependencies {
		if dep.dependency.ArgIndex != nil && *dep.dependency.ArgIndex > maxArgIndex {
			maxArgIndex = *dep.dependency.ArgIndex
		}
	}
	if maxArgIndex != -1 {
		res.ArgList = make([]*string, maxArgIndex+1)
	}

	// build map of default params
	for _, dep := range r.runtimeDependencies {
		if dep.dependency.ArgName != nil {
			res.SetNamedArgVal(dep.value, *dep.dependency.ArgName)

		} else {
			if dep.dependency.ArgIndex == nil {
				return nil, fmt.Errorf("invalid runtime dependency - both ArgName and ArgIndex are nil ")
			}
			res.SetPositionalArgVal(dep.value, *dep.dependency.ArgIndex)
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

// if this leaf run has with runs), execute them
func (r *LeafRun) executeWithRuns(ctx context.Context) {
	for _, w := range r.withRuns {
		go w.Execute(ctx)
	}
	// wait for children to complete
	var errors []error

	for !r.withComplete() {
		log.Printf("[TRACE] run %s waiting for with runs", r.Name)
		completeChild := <-r.childComplete
		log.Printf("[TRACE] run %s got with complete", r.Name)
		if completeChild.GetRunStatus() == dashboardtypes.DashboardRunError {
			errors = append(errors, completeChild.GetError())
		}
		// fall through to recheck ChildrenComplete
	}

	log.Printf("[TRACE] run %s ALL children complete", r.Name)
	// so all with runs have completed - check for errors
	err := error_helpers.CombineErrors(errors...)
	if err == nil {
		if err := r.setWithData(); err != nil {
			r.SetError(ctx, err)
		}
	} else {
		r.SetError(ctx, err)
	}
}

func (r *LeafRun) setWithData() error {
	for _, w := range r.withRuns {
		if err := r.setWithValue(w.DashboardNode.GetUnqualifiedName(), w.Data); err != nil {
			return err
		}
	}
	return nil
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

func (r *LeafRun) getWithValue(name string, path *modconfig.ParsedPropertyPath) (any, error) {
	r.withValueMutex.Lock()
	defer r.withValueMutex.Unlock()

	val, ok := r.withValues[name]
	if !ok {
		return nil, nil
	}

	//  get the set of rows which will be used ot generate the return value
	rows := val.Rows
	/*
			You can
		reference the whole table with:
				with.stuff1
			this is equivalent to:
				with.stuff1.rows
			and
				with.stuff1.rows[*]

			Rows is a list, and you can index it to get a single row:
				with.stuff1.rows[0]
			or splat it to get all rows:
				with.stuff1.rows[*]
			Each row, in turn, contains all the columns, so you can get a single column of a single row:
				with.stuff1.rows[0].a
			if you splat the row, then you can get an array of a single column from all rows. This would be passed to sql as an array:
				with.stuff1.rows[*].a
	*/

	// with.stuff1 -> PropertyPath will be ""
	// with.stuff1.rows -> PropertyPath will be "rows"
	// with.stuff1.rows[*] -> PropertyPath will be "rows.*"
	// with.stuff1.rows[0] -> PropertyPath will be "rows.0"
	// with.stuff1.rows[0].a -> PropertyPath will be "rows.0.a"
	const rowsSegment = 0
	const rowsIdxSegment = 1
	const columnSegment = 2

	// second path section MUST  be "rows"
	if len(path.PropertyPath) > rowsSegment && path.PropertyPath[rowsSegment] != "rows" || len(path.PropertyPath) > (columnSegment+1) {
		return nil, fmt.Errorf("reference to with '%s' has invalid property path '%s'", name, path.Original)
	}

	// if no row is specified assume all
	rowIdxStr := "*"
	if len(path.PropertyPath) > rowsIdxSegment {
		// so there is 3rd part - this will be the row idx (or '*')
		rowIdxStr = path.PropertyPath[rowsIdxSegment]
	}
	var column string

	// is a column specified?
	if len(path.PropertyPath) > columnSegment {
		column = path.PropertyPath[columnSegment]
	} else {
		if len(val.Columns) > 1 {
			// we do not support returning all columns (yet
			return nil, fmt.Errorf("reference to with '%s' is returning more than one column - not supported", name)
		}
		column = val.Columns[0].Name
	}

	if rowIdxStr == "*" {
		return columnValuesFromRows(column, rows)
	}

	rowIdx, err := strconv.Atoi(rowIdxStr)
	if err != nil {
		return nil, fmt.Errorf("reference to with '%s' has invalid property path '%s' - cannot parse row idx '%s'", name, path.Original, rowIdxStr)
	}

	// do we have the requested row
	if rowCount := len(rows); rowIdx >= rowCount {
		return nil, fmt.Errorf("reference to with '%s' has invalid row index '%d' - %d %s were returned", name, rowIdx, rowCount, utils.Pluralize("row", rowCount))
	}
	// so we are returning a single row
	row := rows[rowIdx]
	return row[column], nil

}

func columnValuesFromRows(column string, rows []map[string]interface{}) (any, error) {
	var res = make([]any, len(rows))
	for i, row := range rows {
		var ok bool
		res[i], ok = row[column]
		if !ok {
			return nil, fmt.Errorf("column %s does not exist", column)
		}
	}
	return res, nil
}
func (r *LeafRun) setWithValue(name string, result *dashboardtypes.LeafData) error {
	r.withValueMutex.Lock()
	defer r.withValueMutex.Unlock()

	// TACTICAL - is there are any JSON columns convert them back to a JSON string
	var jsonColumns []string
	for _, c := range result.Columns {
		if c.DataType == "JSONB" || c.DataType == "JSON" {
			jsonColumns = append(jsonColumns, c.Name)
		}
	}
	// now convert any json values into a json string
	for _, c := range jsonColumns {
		for _, row := range result.Rows {
			jsonBytes, err := json.Marshal(row[c])
			if err != nil {
				return err
			}
			row[c] = string(jsonBytes)
		}
	}
	r.withValues[name] = result
	return nil
}
