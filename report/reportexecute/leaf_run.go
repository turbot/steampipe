package reportexecute

import (
	"context"
	"log"

	"github.com/turbot/steampipe/control/controlexecute"
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
	if r.SQL == "" {
		return nil
	}

	var err error
	switch node := r.ReportNode.(type) {
	case *modconfig.ReportControl:
		r.Data, err = r.executeControl(ctx, node)
	default:
		log.Printf("[WARN] %s Execute start", r.Name)
		// if counter has sql execute it
		r.Data, err = r.executeLeafNodeSQL(ctx, r.SQL)
	}
	if err != nil {
		log.Printf("[WARN] %s SQL error %v", r.Name, err)
		// set the error status on the counter - this will raise counter error event
		r.SetError(err)
		return err
	}

	log.Printf("[WARN] %s SetComplete", r.Name)
	// set complete status on counter - this will raise counter complete event
	r.SetComplete()

	log.Printf("[WARN] %s Execute DONE", r.Name)
	return nil
}

func (r *LeafRun) executeLeafNodeSQL(ctx context.Context, query string) (*LeafData, error) {
	log.Printf("[WARN] !!!!!!!!!!!!!!!!!!!!!! EXECUTE SQL START %s !!!!!!!!!!!!!!!!!!!!!!", r.Name)
	queryResult, err := r.executionTree.client.ExecuteSync(ctx, query)
	if err != nil {
		return nil, err
	}
	var res = NewLeafData(queryResult)

	log.Printf("[WARN] $$$$$$$$$$$$$$$$$$ EXECUTE SQL END %s $$$$$$$$$$$$$$$$$$ ", r.Name)

	return res, nil
}

func (r *LeafRun) executeControl(ctx context.Context, reportControl *modconfig.ReportControl) (*LeafData, error) {
	executionTree, err := controlexecute.NewExecutionTree(ctx, r.executionTree.workspace, r.executionTree.client, reportControl.Name())
	if err != nil {
		return nil, err
	}
	executionTree.Execute(ctx)
	return nil, nil
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
	log.Printf("[WARN] **************** LeafRun DONE EVENT %s ***************", r.Name)
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
