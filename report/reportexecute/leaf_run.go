package reportexecute

import (
	"context"

	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// LeafRun is a struct representing the execution of a leaf reporting node
type LeafRun struct {
	Name string `json:"name"`

	Title         string                      `json:"title,omitempty"`
	Width         int                         `json:"width,omitempty"`
	SQL           string                      `json:"sql,omitempty"`
	Data          *LeafData                   `json:"data,omitempty"`
	Error         error                       `json:"error,omitempty"`
	ReportNode    modconfig.ReportingLeafNode `json:"properties"`
	NodeType      string                      `json:"node_type"`
	Path          []string                    `json:"-"`
	parent        reportinterfaces.ReportNodeParent
	runStatus     reportinterfaces.ReportRunStatus
	executionTree *ReportExecutionTree
}

func NewLeafRun(resource modconfig.ReportingLeafNode, parent reportinterfaces.ReportNodeParent, executionTree *ReportExecutionTree) (*LeafRun, error) {
	r := &LeafRun{
		Name:          resource.Name(),
		Title:         resource.GetTitle(),
		Width:         resource.GetWidth(),
		SQL:           resource.GetSQL(),
		Path:          resource.GetPaths()[0],
		ReportNode:    resource,
		executionTree: executionTree,
		parent:        parent,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: reportinterfaces.ReportRunComplete,
	}

	parsedName, err := modconfig.ParseResourceName(resource.Name())
	if err != nil {
		return nil, err
	}
	r.NodeType = parsedName.ItemType
	// if we have sql, set status to ready
	if r.SQL != "" {
		r.runStatus = reportinterfaces.ReportRunReady
	}

	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r, nil
}

// Execute implements ReportRunNode
func (r *LeafRun) Execute(ctx context.Context) error {
	// if there are any unresolved runtime dependencies, wait for them
	if err := r.waitForRuntimeDepdendencies(); err != nil {
		return err
	}

	if r.SQL == "" {
		return nil
	}

	var err error
	queryResult, err := r.executionTree.client.ExecuteSync(ctx, r.SQL)
	if err != nil {
		// set the error status on the counter - this will raise counter error event
		r.SetError(err)
		return err

	}
	r.Data = NewLeafData(queryResult)
	// set complete status on counter - this will raise counter complete event
	r.SetComplete()
	return nil
}

// GetName implements ReportNodeRun
func (r *LeafRun) GetName() string {
	return r.Name
}

// GetPath implements ReportNodeRun
func (r *LeafRun) GetPath() modconfig.NodePath {
	return r.Path
}

// GetRunStatus implements ReportNodeRun
func (r *LeafRun) GetRunStatus() reportinterfaces.ReportRunStatus {
	return r.runStatus
}

// SetError implements ReportNodeRun
func (r *LeafRun) SetError(err error) {
	r.Error = err
	r.runStatus = reportinterfaces.ReportRunError
	// raise counter error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.LeafNodeError{Node: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r

}

// SetComplete implements ReportNodeRun
func (r *LeafRun) SetComplete() {
	r.runStatus = reportinterfaces.ReportRunComplete
	// raise counter complete event
	r.executionTree.workspace.PublishReportEvent(&reportevents.LeafNodeComplete{Node: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements ReportNodeRun
func (r *LeafRun) RunComplete() bool {
	return r.runStatus == reportinterfaces.ReportRunComplete || r.runStatus == reportinterfaces.ReportRunError
}

// ChildrenComplete implements ReportNodeRun
func (r *LeafRun) ChildrenComplete() bool {
	return true
}

func (r *LeafRun) waitForRuntimeDepdendencies() error {
	runtimeDependencies := r.ReportNode.GetRuntimeDependencies()

	for _, v := range runtimeDependencies {
		if err := r.executionTree.waitForRuntimeDependency(v); err != nil {
			return err
		}
	}
	return nil
}
