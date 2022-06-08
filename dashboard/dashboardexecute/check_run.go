package dashboardexecute

import (
	"context"
	"fmt"
	"reflect"

	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/control/controlstatus"
	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// CheckRun is a struct representing the execution of a leaf dashboard node
type CheckRun struct {
	Name             string                             `json:"name"`
	Title            string                             `json:"title,omitempty"`
	Width            int                                `json:"width,omitempty"`
	Description      string                             `json:"description,omitempty"`
	Documentation    string                             `json:"documentation,omitempty"`
	Display          string                             `json:"display,omitempty"`
	Type             string                             `json:"type,omitempty"`
	Tags             map[string]string                  `json:"tags,omitempty"`
	ErrorString      string                             `json:"error,omitempty"`
	NodeType         string                             `json:"node_type"`
	DashboardName    string                             `json:"dashboard"`
	SourceDefinition string                             `json:"source_definition"`
	SessionId        string                             `json:"session_id"`
	Children         []controlexecute.ExecutionTreeNode `json:"children"`
	Summary          *controlexecute.GroupSummary       `json:"summary"`
	// if the dashboard node is a control, serialise to json as 'properties'
	Control       *modconfig.Control          `json:"properties,omitempty"`
	DashboardNode modconfig.DashboardLeafNode `json:"-"`

	controlExecutionTree *controlexecute.ExecutionTree
	error                error

	parent        dashboardinterfaces.DashboardNodeParent
	runStatus     dashboardinterfaces.DashboardRunStatus
	executionTree *DashboardExecutionTree
}

func NewCheckRun(resource modconfig.DashboardLeafNode, parent dashboardinterfaces.DashboardNodeParent, executionTree *DashboardExecutionTree) (*CheckRun, error) {

	// NOTE: for now we MUST declare container/dashboard children inline - therefore we cannot share children between runs in the tree
	// (if we supported the children property then we could reuse resources)
	// so FOR NOW it is safe to use the node name directly as the run name
	name := resource.Name()

	c := &CheckRun{
		Name:             name,
		Title:            resource.GetTitle(),
		Width:            resource.GetWidth(),
		Description:      resource.GetDescription(),
		Documentation:    resource.GetDocumentation(),
		Display:          resource.GetDisplay(),
		Type:             resource.GetType(),
		Tags:             resource.GetTags(),
		DashboardName:    executionTree.dashboardName,
		SourceDefinition: resource.GetMetadata().SourceDefinition,
		SessionId:        executionTree.sessionId,
		executionTree:    executionTree,
		DashboardNode:    resource,
		parent:           parent,

		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		runStatus: dashboardinterfaces.DashboardRunComplete,
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
	c.runStatus = dashboardinterfaces.DashboardRunReady

	// add r into execution tree
	executionTree.runs[c.Name] = c
	return c, nil
}

// Initialise implements DashboardRunNode
func (r *CheckRun) Initialise(ctx context.Context) {
	// build control execution tree during init, rather than in Execute, so that it is populated when the ExecutionStarted event is sent
	executionTree, err := controlexecute.NewExecutionTree(ctx, r.executionTree.workspace, r.executionTree.client, r.DashboardNode.Name())
	if err != nil {
		// set the error status on the counter - this will raise counter error event
		r.SetError(err)
		return
	}
	r.controlExecutionTree = executionTree
	// if we are executing a benchmark, set children
	if rootBenchmark, ok := executionTree.Root.Children[0].(*controlexecute.ResultGroup); ok {
		r.Children = rootBenchmark.Children
	}
}

// Execute implements DashboardRunNode
func (r *CheckRun) Execute(ctx context.Context) {

	// create a context with a ControlEventHooks to report control execution progress
	ctx = controlstatus.AddControlHooksToContext(ctx, NewControlEventHooks(r))
	r.controlExecutionTree.Execute(ctx)

	// set the summary on the CeckRun
	r.Summary = r.controlExecutionTree.Root.Summary

	// set complete status on counter - this will raise counter complete event
	r.SetComplete()

}

// GetName implements DashboardNodeRun
func (r *CheckRun) GetName() string {
	return r.Name
}

// GetRunStatus implements DashboardNodeRun
func (r *CheckRun) GetRunStatus() dashboardinterfaces.DashboardRunStatus {
	return r.runStatus
}

// SetError implements DashboardNodeRun
func (r *CheckRun) SetError(err error) {
	r.error = err
	// error type does not serialise to JSON so copy into a string
	r.ErrorString = err.Error()

	r.runStatus = dashboardinterfaces.DashboardRunError
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

// GetError implements DashboardNodeRun
func (r *CheckRun) GetError() error {
	return r.error
}

// SetComplete implements DashboardNodeRun
func (r *CheckRun) SetComplete() {
	r.runStatus = dashboardinterfaces.DashboardRunComplete
	// raise counter complete event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeComplete{
		LeafNode:    r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
	})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements DashboardNodeRun
func (r *CheckRun) RunComplete() bool {
	return r.runStatus == dashboardinterfaces.DashboardRunComplete || r.runStatus == dashboardinterfaces.DashboardRunError
}

// ChildrenComplete implements DashboardNodeRun
func (r *CheckRun) ChildrenComplete() bool {
	return r.RunComplete()
}

// GetInputsDependingOn implements DashboardNodeRun
//return nothing for CheckRun
func (r *CheckRun) GetInputsDependingOn(changedInputName string) []string { return nil }
