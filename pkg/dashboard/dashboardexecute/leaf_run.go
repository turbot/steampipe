package dashboardexecute

import (
	"context"
	"fmt"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/type_conversion"
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
	Args             []string                          `json:"args,omitempty"`
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
}

func (r *LeafRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	return &dashboardtypes.SnapshotTreeNode{
		Name:     r.Name,
		NodeType: r.NodeType,
	}
}

func NewLeafRun(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardNodeParent, executionTree *DashboardExecutionTree) (*LeafRun, error) {
	// NOTE: for now we MUST declare container/dashboard children inline - therefore we cannot share children between runs in the tree
	// (if we supported the children property then we could reuse resources)
	// so FOR NOW it is safe to use the node name directly as the run name
	name := resource.Name()

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
		withValues:          make(map[string]*dashboardtypes.LeafData),
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
	if children := resource.GetChildren(); len(children) > 0 {
		// create the child runs
		return r.createChildRuns(children, executionTree)
	}
	return r, nil
}

// if this node has runtime dependencies, create runtime dependency instances which we use to resolve the values
func (r *LeafRun) addRuntimeDependencies() {
	// only QueryProvider resources support runtime dependencies
	queryProvider, ok := r.DashboardNode.(modconfig.QueryProvider)
	if !ok {
		return
	}
	runtimeDependencies := queryProvider.GetRuntimeDependencies()
	for name, dep := range runtimeDependencies {
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
	if parentLeafRun, ok := r.parent.(*LeafRun); ok {
		for name, dep := range parentLeafRun.runtimeDependencies {
			if _, ok := r.runtimeDependencies[name]; !ok {
				r.runtimeDependencies[name] = dep
			}
		}
	}
}

func (r *LeafRun) createChildRuns(children []modconfig.ModTreeItem, executionTree *DashboardExecutionTree) (*LeafRun, error) {
	// create buffered child complete chan
	r.childComplete = make(chan dashboardtypes.DashboardNodeRun, len(children))

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
	return r, error_helpers.CombineErrors(errors...)
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
	for _, w := range r.DashboardNode.(modconfig.QueryProvider).GetWiths() {
		queryResult, err := r.executionTree.client.ExecuteSync(ctx, typehelpers.SafeString(w.SQL))
		if err != nil {
			r.SetError(ctx, err)
			return
		}
		withResult := dashboardtypes.NewLeafData(queryResult)
		r.setWithValue(w.UnqualifiedName, withResult)
	}

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
		// format the arg value as a postgres string
		formattedVal, err := type_conversion.GoToPostgresString(dep.value)
		if err != nil {
			return nil, err
		}

		if dep.dependency.ArgName != nil {
			res.ArgMap[*dep.dependency.ArgName] = formattedVal
		} else {
			if dep.dependency.ArgIndex == nil {
				return nil, fmt.Errorf("invalid runtime dependency - both ArgName and ArgIndex are nil ")
			}

			// now add at correct index
			res.ArgList[*dep.dependency.ArgIndex] = &formattedVal
		}
	}
	return res, nil
}

// if this leaf run has a query or sql, execute it now
func (r *LeafRun) executeQuery(ctx context.Context) {
	log.Printf("[TRACE] LeafRun '%s' SQL resolved, executing", r.DashboardNode.Name())

	queryResult, err := r.executionTree.client.ExecuteSync(ctx, r.executeSQL)
	if err != nil {
		query := r.DashboardNode.(modconfig.QueryProvider).GetQuery()
		if query != nil {
			queryName := query.Name()
			// get the query and any prepared statement error from the workspace
			preparedStatementFailure := r.executionTree.workspace.GetPreparedStatementCreationFailure(queryName)
			if preparedStatementFailure != nil {
				declRange := preparedStatementFailure.Query.GetDeclRange()
				preparedStatementError := preparedStatementFailure.Error
				err = error_helpers.EnrichPreparedStatementError(err, queryName, preparedStatementError, declRange)
			}
		}
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
	err := error_helpers.CombineErrors(errors...)
	if err == nil {
		r.combineChildData()
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
		You can reference the whole table with:
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
		return nil, fmt.Errorf("with '%s' has invalid property path '%s'", name, path.Original)
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
			return nil, fmt.Errorf("with '%s' is returning more than one column - not supported", name)
		}
		column = val.Columns[0].Name
	}

	if rowIdxStr == "*" {
		return columnValuesFromRows(column, rows)
	}

	rowIdx, err := strconv.Atoi(rowIdxStr)
	if err != nil {
		return nil, fmt.Errorf("with '%s' has invalid property path '%s' - cannot parse row idx '%s'", name, path.Original, rowIdxStr)
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

func (r *LeafRun) setWithValue(name string, result *dashboardtypes.LeafData) {
	r.withValueMutex.Lock()
	defer r.withValueMutex.Unlock()

	r.withValues[name] = result
}
