package reportexecute

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// CheckRun is a struct representing the execution of a leaf reporting node
type CheckRun struct {
	Name string `json:"name"`

	Title                string                        `json:"title,omitempty"`
	Width                int                           `json:"width,omitempty"`
	Error                error                         `json:"error,omitempty"`
	NodeType             string                        `json:"node_type"`
	ControlExecutionTree *controlexecute.ExecutionTree `json:"execution_tree"`
	ReportNode           modconfig.ReportingLeafNode   `json:"-"`
	Path                 []string                      `json:"-"`
	parent               reportinterfaces.ReportNodeParent
	runStatus            reportinterfaces.ReportRunStatus
	executionTree        *ReportExecutionTree
}

func NewCheckRun(resource modconfig.ReportingLeafNode, parent reportinterfaces.ReportNodeParent, executionTree *ReportExecutionTree) (*CheckRun, error) {
	r := &CheckRun{
		Name:          resource.Name(),
		Title:         resource.GetTitle(),
		Width:         resource.GetWidth(),
		Path:          resource.GetPaths()[0],
		ReportNode:    resource,
		executionTree: executionTree,
		parent:        parent,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: reportinterfaces.ReportRunComplete,
	}
	// verify node type
	switch resource.(type) {
	case *modconfig.Control:
		r.NodeType = modconfig.BlockTypeControl
	case *modconfig.Benchmark:
		r.NodeType = modconfig.BlockTypeBenchmark
	default:
		return nil, fmt.Errorf("check run instantiated with invalid node type %s", reflect.TypeOf(resource).Name())
	}

	//  set status to ready
	r.runStatus = reportinterfaces.ReportRunReady

	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r, nil
}

// Execute implements ReportRunNode
func (r *CheckRun) Execute(ctx context.Context) error {
	executionTree, err := controlexecute.NewExecutionTree(ctx, r.executionTree.workspace, r.executionTree.client, r.ReportNode.Name())
	if err != nil {
		log.Printf("[WARN] %s Control execution error %v", r.Name, err)
		// set the error status on the counter - this will raise counter error event
		r.SetError(err)
		return err
	}
	executionTree.Execute(ctx)
	r.ControlExecutionTree = executionTree

	log.Printf("[WARN] %s SetComplete", r.Name)
	// set complete status on counter - this will raise counter complete event
	r.SetComplete()

	log.Printf("[WARN] %s Execute DONE", r.Name)
	return nil
}

// GetName implements ReportNodeRun
func (r *CheckRun) GetName() string {
	return r.Name
}

// GetPath implements ReportNodeRun
func (r *CheckRun) GetPath() modconfig.NodePath {
	return r.Path
}

// GetRunStatus implements ReportNodeRun
func (r *CheckRun) GetRunStatus() reportinterfaces.ReportRunStatus {
	return r.runStatus
}

// SetError implements ReportNodeRun
func (r *CheckRun) SetError(err error) {
	r.Error = err
	r.runStatus = reportinterfaces.ReportRunError
	// raise counter error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.LeafNodeError{Node: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r

}

// SetComplete implements ReportNodeRun
func (r *CheckRun) SetComplete() {
	r.runStatus = reportinterfaces.ReportRunComplete
	// raise counter complete event
	log.Printf("[WARN] **************** CheckRun DONE EVENT %s ***************", r.Name)
	r.executionTree.workspace.PublishReportEvent(&reportevents.LeafNodeComplete{Node: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements ReportNodeRun
func (r *CheckRun) RunComplete() bool {
	return r.runStatus == reportinterfaces.ReportRunComplete || r.runStatus == reportinterfaces.ReportRunError
}

// ChildrenComplete implements ReportNodeRun
func (r *CheckRun) ChildrenComplete() bool {
	return r.RunComplete()
}
