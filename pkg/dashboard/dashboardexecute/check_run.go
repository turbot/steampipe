package dashboardexecute

import (
	"context"

	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

// CheckRun is a struct representing the execution of a control or benchmark
type CheckRun struct {
	DashboardParentImpl

	Summary   *controlexecute.GroupSummary `json:"summary"`
	SessionId string                       `json:"-"`
	// if the dashboard node is a control, serialise to json as 'properties'
	Control *modconfig.Control               `json:"properties,omitempty"`
	Root    controlexecute.ExecutionTreeNode `json:"-"`

	controlExecutionTree *controlexecute.ExecutionTree
}

func (r *CheckRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	return r.Root.AsTreeNode()
}

func NewCheckRun(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) (*CheckRun, error) {
	c := &CheckRun{SessionId: executionTree.sessionId}
	// create NewDashboardTreeRunImpl
	// (we must create after creating the run as it requires a ref to the run)
	c.DashboardParentImpl = newDashboardParentImpl(resource, parent, c, executionTree)

	c.NodeType = resource.BlockType()
	//  set status to initialized
	c.Status = dashboardtypes.RunInitialized
	// add r into execution tree
	executionTree.runs[c.Name] = c
	return c, nil
}

// Initialise implements DashboardTreeRun
func (r *CheckRun) Initialise(ctx context.Context) {
	// build control execution tree during init, rather than in Execute, so that it is populated when the ExecutionStarted event is sent
	controlFilterWhereClause := ""
	executionTree, err := controlexecute.NewExecutionTree(ctx, r.executionTree.workspace, r.executionTree.client, controlFilterWhereClause, r.resource.Name())
	if err != nil {
		// set the error status on the counter - this will raise counter error event
		r.SetError(ctx, err)
		return
	}
	r.controlExecutionTree = executionTree
	r.Root = executionTree.Root.Children[0]
}

// Execute implements DashboardTreeRun
func (r *CheckRun) Execute(ctx context.Context) {
	utils.LogTime("CheckRun.execute start")
	defer utils.LogTime("CheckRun.execute end")

	// set status to running (this sends update event)
	r.setRunning(ctx)

	// create a context with a DashboardEventControlHooks to report control execution progress
	ctx = controlstatus.AddControlHooksToContext(ctx, NewDashboardEventControlHooks(r))
	if err := r.controlExecutionTree.Execute(ctx); err != nil {
		r.SetError(ctx, err)
		return
	}

	// set the summary on the CheckRun
	r.Summary = r.controlExecutionTree.Root.Summary

	// set complete status on counter - this will raise counter complete event
	r.SetComplete(ctx)
}

// ChildrenComplete implements DashboardTreeRun (override base)
func (r *CheckRun) ChildrenComplete() bool {
	return r.RunComplete()
}

// IsSnapshotPanel implements SnapshotPanel
func (*CheckRun) IsSnapshotPanel() {}

// SetError implements DashboardTreeRun (override to set snapshothook status
func (r *CheckRun) SetError(ctx context.Context, err error) {
	// increment error count for snapshot hook
	statushooks.SnapshotError(ctx)
	r.DashboardTreeRunImpl.SetError(ctx, err)
}

// SetComplete implements DashboardTreeRun (override to set snapshothook status
func (r *CheckRun) SetComplete(ctx context.Context) {
	// call snapshot hooks with progress
	statushooks.UpdateSnapshotProgress(ctx, 1)

	r.DashboardTreeRunImpl.SetComplete(ctx)
}

// BuildSnapshotPanels is a custom implementation of BuildSnapshotPanels - be nice to just use the DashboardExecutionTree but work is needed on common interface types/generics
func (r *CheckRun) BuildSnapshotPanels(leafNodeMap map[string]dashboardtypes.SnapshotPanel) map[string]dashboardtypes.SnapshotPanel {
	// if this check run is for a control, just add the controlRUn
	if controlRun, ok := r.Root.(*controlexecute.ControlRun); ok {
		leafNodeMap[controlRun.Control.Name()] = controlRun
		return leafNodeMap
	}

	leafNodeMap[r.GetName()] = r

	return r.buildSnapshotPanelsUnder(r.Root, leafNodeMap)
}

func (r *CheckRun) buildSnapshotPanelsUnder(parent controlexecute.ExecutionTreeNode, res map[string]dashboardtypes.SnapshotPanel) map[string]dashboardtypes.SnapshotPanel {
	for _, c := range parent.GetChildren() {
		// if this node is a snapshot node, add to map
		if snapshotNode, ok := c.(dashboardtypes.SnapshotPanel); ok {
			res[c.GetName()] = snapshotNode
		}
		res = r.buildSnapshotPanelsUnder(c, res)
	}
	return res
}
