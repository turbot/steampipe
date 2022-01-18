package reportexecute

import (
	"context"
	"log"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// PanelRun is a struct representing a panel run
type PanelRun struct {
	Name       string            `json:"name"`
	Title      string            `json:"title,omitempty"`
	Type       string            `json:"type,omitempty"`
	Width      int               `json:"width,omitempty"`
	SQL        string            `json:"sql,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`

	Data  [][]interface{} `json:"data,omitempty"`
	Error error           `json:"error,omitempty"`

	parent        reportinterfaces.ReportNodeParent
	runStatus     reportinterfaces.ReportRunStatus
	executionTree *ReportExecutionTree
}

func NewPanelRun(panel *modconfig.Panel, parent reportinterfaces.ReportNodeParent, executionTree *ReportExecutionTree) *PanelRun {
	r := &PanelRun{
		Name:          panel.Name(),
		Title:         typehelpers.SafeString(panel.Title),
		Properties:    panel.Properties,
		Type:          typehelpers.SafeString(panel.Type),
		SQL:           typehelpers.SafeString(panel.SQL),
		executionTree: executionTree,
		parent:        parent,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: reportinterfaces.ReportRunComplete,
	}
	if panel.Width != nil {
		r.Width = *panel.Width
	}

	// if we have sql, set status to ready
	if panel.SQL != nil {
		r.runStatus = reportinterfaces.ReportRunReady
	}

	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r
}

// Execute implements ReportRunNode
func (r *PanelRun) Execute(ctx context.Context) error {
	log.Printf("[WARN] %s Execute start", r.Name)
	// if panel has sql execute it
	if r.SQL != "" {
		data, err := r.executePanelSQL(ctx, r.SQL)
		if err != nil {
			log.Printf("[WARN] %s SQL error %v", r.Name, err)
			// set the error status on the panel - this will raise panel error event
			r.SetError(err)
			return err
		}

		r.Data = data
		log.Printf("[WARN] %s SetComplete", r.Name)
		// set complete status on panel - this will raise panel complete event
		r.SetComplete()
	}
	log.Printf("[WARN] %s Execute DONE", r.Name)
	return nil
}

func (r *PanelRun) executePanelSQL(ctx context.Context, query string) ([][]interface{}, error) {
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
func (r *PanelRun) GetName() string {
	return r.Name
}

// GetRunStatus implements ReportNodeRun
func (r *PanelRun) GetRunStatus() reportinterfaces.ReportRunStatus {
	return r.runStatus
}

// SetError implements ReportNodeRun
func (r *PanelRun) SetError(err error) {
	r.Error = err
	r.runStatus = reportinterfaces.ReportRunError
	// raise panel error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.PanelError{Panel: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r

}

// SetComplete implements ReportNodeRun
func (r *PanelRun) SetComplete() {
	r.runStatus = reportinterfaces.ReportRunComplete
	// raise panel complete event
	log.Printf("[WARN] **************** PANEL DONE EVENT %s ***************", r.Name)
	r.executionTree.workspace.PublishReportEvent(&reportevents.PanelComplete{Panel: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements ReportNodeRun
func (r *PanelRun) RunComplete() bool {
	return r.runStatus == reportinterfaces.ReportRunComplete || r.runStatus == reportinterfaces.ReportRunError
}

// ChildrenComplete implements ReportNodeRun
func (r *PanelRun) ChildrenComplete() bool {
	return true
}
