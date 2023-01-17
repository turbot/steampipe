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
	RuntimeDependencySubscriberImpl
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

func NewLeafRun(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree, opts ...LeafRunOption) (*LeafRun, error) {
	r := &LeafRun{
		Resource: resource,
	}

	// create RuntimeDependencySubscriberImpl- this handles 'with' run creation and resolving runtime dependency resolution
	// (NOTE: we have to do this after creating run as we need to pass a ref to the run)
	r.RuntimeDependencySubscriberImpl = *NewRuntimeDependencySubscriber(resource, parent, r, executionTree)

	// apply options AFTER calling NewRuntimeDependencySubscriber
	for _, opt := range opts {
		opt(r)
	}

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

	for i, c := range children {
		var opts []LeafRunOption
		childRun, err := NewLeafRun(c.(modconfig.DashboardLeafNode), r, executionTree, opts...)
		if err != nil {
			errors = append(errors, err)
			continue
		}

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
	if r.Status.IsFinished() {
		return
	}

	log.Printf("[TRACE] LeafRun '%s' Execute()", r.resource.Name())

	// to get here, we must be a query provider

	// if we have children and with runs, start them asyncronously (they may block waiting for our runtime dependencies)
	r.executeChildrenAsync(ctx)

	// start a goroutine to wait for children to complete
	doneChan := r.waitForChildrenAsync(ctx)

	if err := r.evaluateRuntimeDependencies(ctx); err != nil {
		r.SetError(ctx, err)
		return
	}

	// set status to running (this sends update event)
	// (if we have blocked children, this will be changed to blocked)
	r.setRunning(ctx)

	// if we have sql to execute, do it now
	// (if we are only performing a base execution, do not run the query)
	if r.executeSQL != "" {
		if err := r.executeQuery(ctx); err != nil {
			r.SetError(ctx, err)
			return
		}
	}

	// wait for all children and withs
	err := <-doneChan
	if err == nil {
		log.Printf("[TRACE] %s children complete", r.resource.Name())

		// aggregate our child data
		r.combineChildData()
		// set complete status on dashboard
		r.SetComplete(ctx)
	} else {
		log.Printf("[TRACE] %s children complete with error: %s", r.resource.Name(), err.Error())
		r.SetError(ctx, err)
	}
}

// SetError implements DashboardTreeRun (override to set snapshothook status)
func (r *LeafRun) SetError(ctx context.Context, err error) {
	// increment error count for snapshot hook
	statushooks.SnapshotError(ctx)
	r.DashboardTreeRunImpl.SetError(ctx, err)
}

// SetComplete implements DashboardTreeRun (override to set snapshothook status
func (r *LeafRun) SetComplete(ctx context.Context) {
	// call snapshot hooks with progress
	statushooks.UpdateSnapshotProgress(ctx, 1)

	r.DashboardTreeRunImpl.SetComplete(ctx)
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
