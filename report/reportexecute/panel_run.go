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

// CounterRun is a struct representing a counter run
type CounterRun struct {
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
	Type  string `json:"type,omitempty"`
	Width int    `json:"width,omitempty"`
	SQL   string `json:"sql,omitempty"`

	Data  [][]interface{} `json:"data,omitempty"`
	Error error           `json:"error,omitempty"`

	parent        reportinterfaces.ReportNodeParent
	runStatus     reportinterfaces.ReportRunStatus
	executionTree *ReportExecutionTree
}

func NewCounterRun(counter *modconfig.ReportCounter, parent reportinterfaces.ReportNodeParent, executionTree *ReportExecutionTree) *CounterRun {
	r := &CounterRun{
		Name:          counter.Name(),
		Title:         typehelpers.SafeString(counter.Title),
		Type:          typehelpers.SafeString(counter.Type),
		SQL:           typehelpers.SafeString(counter.SQL),
		executionTree: executionTree,
		parent:        parent,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: reportinterfaces.ReportRunComplete,
	}
	if counter.Width != nil {
		r.Width = *counter.Width
	}

	// if we have sql, set status to ready
	if counter.SQL != nil {
		r.runStatus = reportinterfaces.ReportRunReady
	}

	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r
}

// Execute implements ReportRunNode
func (r *CounterRun) Execute(ctx context.Context) error {
	log.Printf("[WARN] %s Execute start", r.Name)
	// if counter has sql execute it
	if r.SQL != "" {
		data, err := r.executeCounterSQL(ctx, r.SQL)
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

func (r *CounterRun) executeCounterSQL(ctx context.Context, query string) ([][]interface{}, error) {
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
func (r *CounterRun) GetName() string {
	return r.Name
}

// GetRunStatus implements ReportNodeRun
func (r *CounterRun) GetRunStatus() reportinterfaces.ReportRunStatus {
	return r.runStatus
}

// SetError implements ReportNodeRun
func (r *CounterRun) SetError(err error) {
	r.Error = err
	r.runStatus = reportinterfaces.ReportRunError
	// raise counter error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.CounterError{Counter: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r

}

// SetComplete implements ReportNodeRun
func (r *CounterRun) SetComplete() {
	r.runStatus = reportinterfaces.ReportRunComplete
	// raise counter complete event
	log.Printf("[WARN] **************** COUNTER DONE EVENT %s ***************", r.Name)
	r.executionTree.workspace.PublishReportEvent(&reportevents.CounterComplete{Counter: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements ReportNodeRun
func (r *CounterRun) RunComplete() bool {
	return r.runStatus == reportinterfaces.ReportRunComplete || r.runStatus == reportinterfaces.ReportRunError
}

// ChildrenComplete implements ReportNodeRun
func (r *CounterRun) ChildrenComplete() bool {
	return true
}
