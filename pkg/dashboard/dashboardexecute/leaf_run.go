package dashboardexecute

import (
	"context"
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
	// all RuntimeDependencySubscribers are also publishers as they have args/params
	RuntimeDependencySubscriber
	Resource modconfig.DashboardLeafNode `json:"properties,omitempty"`

	Data         *dashboardtypes.LeafData  `json:"data,omitempty"`
	TimingResult *queryresult.TimingResult `json:"-"`
	// function called when the run is complete
	// this property populated for 'with' runs
	onComplete func()
}

func (r *LeafRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	return &dashboardtypes.SnapshotTreeNode{
		Name:     r.Name,
		NodeType: r.NodeType,
	}
}

func NewLeafRun(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) (*LeafRun, error) {
	r := &LeafRun{
		Resource: resource,
	}

	// create RuntimeDependencySubscriber- this handles 'with' run creation and resolving runtime dependency resolution
	// (NOTE: we have to do this after creating run as we need to pass a ref to the run)
	r.RuntimeDependencySubscriber = *NewRuntimeDependencySubscriber(resource, parent, r, executionTree)

	err := r.initRuntimeDependencies(executionTree)
	if err != nil {
		return nil, err
	}

	r.NodeType = resource.BlockType()

	// if the node has no runtime dependencies, resolve the sql
	if !r.hasRuntimeDependencies() {
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
	r.setRuntimeDependencies()

	return r, nil
}

func (r *LeafRun) createChildRuns(executionTree *DashboardExecutionTree) error {
	children := r.resource.GetChildren()
	if len(children) == 0 {
		return nil
	}

	r.children = make([]dashboardtypes.DashboardTreeRun, len(children))
	var errors []error

	// if the leaf run has children (nodes/edges) create runs for them
	inheritedChildren := r.resource.(modconfig.NodeAndEdgeProvider).GetInheritedChildren()

	for i, c := range children {
		// TACTICAL if nodes/edges have been inherited from a base NodeEdgeProvider resource,
		// create the run passing the BASE resource as the parent
		// this ensures we resolve runtime dependencies from the base resource
		var parent dashboardtypes.DashboardParent = r
		isInherited := inheritedChildren[c.Name()]
		if isInherited {
			// set parent to the base run
			parent = r.baseDependencySubscriber
		}

		childRun, err := NewLeafRun(c.(modconfig.DashboardLeafNode), parent, executionTree)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		// TACTICAL reset parent
		r.parent = r

		r.children[i] = childRun
	}
	return error_helpers.CombineErrors(errors...)
}

// Execute implements DashboardTreeRun
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

	log.Printf("[TRACE] LeafRun '%s' Execute()", r.resource.Name())

	// to get here, we must be a query provider

	// if we have children and with runs, start them asyncronously (they may block waiting for our runtime dependencies)
	r.executeChildrenAsync(ctx)

	// start a goroutine to wait for children to complete
	doneChan := r.waitForChildrenAsync()

	if err := r.evaluateRuntimeDependencies(); err != nil {
		r.SetError(ctx, err)
		return
	}

	// set status to running (this sends update event)
	r.setStatus(dashboardtypes.DashboardRunRunning)

	// if we have sql to execute, do it now
	// (if we are only performing a base execution, do not run the query)
	if r.executeSQL != "" && !r.executeConfig.BaseExecution {
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

	// increment error count for snapshot hook
	statushooks.SnapshotError(ctx)
	// set status (this sends update event)
	r.setStatus(dashboardtypes.DashboardRunError)

	r.notifyParentOfCompletion()
}

// SetComplete implements DashboardTreeRun
func (r *LeafRun) SetComplete(ctx context.Context) {
	// set status (this sends update event)
	r.setStatus(dashboardtypes.DashboardRunComplete)

	// call snapshot hooks with progress
	statushooks.UpdateSnapshotProgress(ctx, 1)

	// tell parent we are done
	r.notifyParentOfCompletion()
}

// IsSnapshotPanel implements SnapshotPanel
func (*LeafRun) IsSnapshotPanel() {}

// if this leaf run has a query or sql, execute it now
func (r *LeafRun) executeQuery(ctx context.Context) error {
	log.Printf("[TRACE] LeafRun '%s' SQL resolved, executing", r.resource.Name())

	queryResult, err := r.executionTree.client.ExecuteSync(ctx, r.executeSQL, r.Args...)
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' query failed: %s", r.resource.Name(), err.Error())
		return err

	}
	log.Printf("[TRACE] LeafRun '%s' complete", r.resource.Name())

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
		if data == nil || childLeafRun.resource.BlockType() == modconfig.BlockTypeWith {
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
