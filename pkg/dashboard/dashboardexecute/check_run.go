package dashboardexecute

import (
	"context"
	"fmt"
	"reflect"

	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

// CheckRun is a struct representing the execution of a control or benchmark
type CheckRun struct {
	DashboardParentImpl

	Summary   *controlexecute.GroupSummary `json:"summary"`
	SessionId string                       `json:"-"`
	// if the dashboard node is a control, serialise to json as 'properties'
	Control       *modconfig.Control               `json:"properties,omitempty"`
	DashboardNode modconfig.DashboardLeafNode      `json:"-"`
	Root          controlexecute.ExecutionTreeNode `json:"-"`

	controlExecutionTree *controlexecute.ExecutionTree
	parent               dashboardtypes.DashboardParent
	runStatus            dashboardtypes.DashboardRunStatus
	executionTree        *DashboardExecutionTree
}

func (r *CheckRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	return r.Root.AsTreeNode()
}

func NewCheckRun(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) (*CheckRun, error) {
	c := &CheckRun{
		DashboardParentImpl: DashboardParentImpl{
			DashboardTreeRunImpl: NewDashboardTreeRunImpl(resource, executionTree, executionTree),
		},

		SessionId:     executionTree.sessionId,
		executionTree: executionTree,
		DashboardNode: resource,
		parent:        parent,

		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		runStatus: dashboardtypes.DashboardRunComplete,
	}
	// verify node type
	switch t := resource.(type) {
	case *modconfig.Control:
		c.NodeType = modconfig.BlockTypeControl
		c.Control = t
	case *modconfig.Benchmark:
		c.NodeType = modconfig.BlockTypeBenchmark
	default:
		return nil, fmt.Errorf("check run instantiated with invalid node type %s", reflect.TypeOf(resource).Name())
	}

	//  set status to ready
	c.runStatus = dashboardtypes.DashboardRunReady

	// add r into execution tree
	executionTree.runs[c.Name] = c
	return c, nil
}

// Initialise implements DashboardRunNode
func (r *CheckRun) Initialise(ctx context.Context) {
	// build control execution tree during init, rather than in Execute, so that it is populated when the ExecutionStarted event is sent
	executionTree, err := controlexecute.NewExecutionTree(ctx, r.executionTree.workspace, r.executionTree.client, r.DashboardNode.Name(), "")
	if err != nil {
		// set the error status on the counter - this will raise counter error event
		r.SetError(ctx, err)
		return
	}
	r.controlExecutionTree = executionTree
	r.Root = executionTree.Root.Children[0]
}

// Execute implements DashboardRunNode
func (r *CheckRun) Execute(ctx context.Context) {
	utils.LogTime("CheckRun.execute start")
	defer utils.LogTime("CheckRun.execute end")

	// create a context with a DashboardEventControlHooks to report control execution progress
	ctx = controlstatus.AddControlHooksToContext(ctx, NewDashboardEventControlHooks(r))
	r.controlExecutionTree.Execute(ctx)

	// set the summary on the CheckRun
	r.Summary = r.controlExecutionTree.Root.Summary

	// set complete status on counter - this will raise counter complete event
	r.SetComplete(ctx)
}

// GetRunStatus implements DashboardTreeRun
func (r *CheckRun) GetRunStatus() dashboardtypes.DashboardRunStatus {
	return r.runStatus
}

// SetError implements DashboardTreeRun
func (r *CheckRun) SetError(ctx context.Context, err error) {
	r.err = err
	// error type does not serialise to JSON so copy into a string
	r.ErrorString = err.Error()

	r.runStatus = dashboardtypes.DashboardRunError
	// raise dashboard error event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeError{
		LeafNode:    r,
		Session:     r.executionTree.sessionId,
		Error:       err,
		ExecutionId: r.executionTree.id,
	})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// SetComplete implements DashboardTreeRun
func (r *CheckRun) SetComplete(ctx context.Context) {
	r.runStatus = dashboardtypes.DashboardRunComplete
	// raise counter complete event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeComplete{
		LeafNode:    r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
	})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// ChildrenComplete implements DashboardTreeRun (override base)
func (r *CheckRun) ChildrenComplete() bool {
	return r.RunComplete()
}

// IsSnapshotPanel implements SnapshotPanel
func (*CheckRun) IsSnapshotPanel() {}

// GetTitle implements DashboardTreeRun
func (r *CheckRun) GetTitle() string {
	return r.Title
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
