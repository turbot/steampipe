package reportexecute

import (
	"context"
	"log"

	"github.com/turbot/steampipe/query/queryresult"
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
	Data          [][]interface{}             `json:"data,omitempty"`
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
	// todo check whether leafnode has an execute function

	log.Printf("[WARN] %s Execute start", r.Name)
	// if counter has sql execute it
	if r.SQL != "" {
		data, err := r.executeLeafNodeSQL(ctx, r.SQL)
		if err != nil {
			log.Printf("[WARN] %s SQL error %v", r.Name, err)
			// set the error status on the counter - this will raise counter error event
			r.SetError(err)
			return err
		}

		r.Data = data
		log.Printf("[WARN] %s SetComplete", r.Name)
		// set complete status on counter - this will raise counter complete event
		r.SetComplete()
	}
	log.Printf("[WARN] %s Execute DONE", r.Name)
	return nil
}

func (r *LeafRun) executeLeafNodeSQL(ctx context.Context, query string) ([][]interface{}, error) {
	log.Printf("[WARN] !!!!!!!!!!!!!!!!!!!!!! EXECUTE SQL START %s !!!!!!!!!!!!!!!!!!!!!!", r.Name)
	queryResult, err := r.executionTree.client.ExecuteSync(ctx, query)
	if err != nil {
		return nil, err
	}
	var res = make([][]interface{}, len(queryResult.Rows)+1)
	var columns = make([]interface{}, len(queryResult.ColTypes))
	for i, c := range queryResult.ColTypes {
		columns[i] = c.Name()
	}
	res[0] = columns
	for i, row := range queryResult.Rows {
		rowData := make([]interface{}, len(queryResult.ColTypes))
		for j, columnVal := range row.(*queryresult.RowResult).Data {
			rowData[j] = columnVal
		}
		res[i+1] = rowData
	}

	log.Printf("[WARN] $$$$$$$$$$$$$$$$$$ EXECUTE SQL END %s $$$$$$$$$$$$$$$$$$ ", r.Name)

	return res, nil
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
