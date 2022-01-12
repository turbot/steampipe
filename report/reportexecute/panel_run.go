package reportexecute

import (
	"fmt"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// PanelRun is a struct representing a panel run
type PanelRun struct {
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
	Type  string `json:"type,omitempty"`
	Width int    `json:"width,omitempty"`
	SQL   string `json:"sql,omitempty"`
	Text  string `json:"text,omitempty"`

	Properties map[string]string

	Data  [][]interface{} `json:"data,omitempty"`
	Error error           `json:"error,omitempty"`

	runStatus     reportinterfaces.ReportRunStatus
	executionTree *ReportExecutionTree
}

func NewPanelRun(panel *modconfig.Panel, parentName string, executionTree *ReportExecutionTree) *PanelRun {
	r := &PanelRun{
		// the name is the path, i.e. dot-separated concatenation of parent names
		Name:          fmt.Sprintf("%s.%s", parentName, panel.UnqualifiedName),
		Title:         typehelpers.SafeString(panel.Title),
		Properties:    panel.Properties,
		Type:          typehelpers.SafeString(panel.Type),
		SQL:           typehelpers.SafeString(panel.SQL),
		executionTree: executionTree,

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

}

// SetComplete implements ReportNodeRun
func (r *PanelRun) SetComplete() {
	r.runStatus = reportinterfaces.ReportRunComplete
	// raise panel complete event
	r.executionTree.workspace.PublishReportEvent(&reportevents.PanelComplete{Panel: r})
}

// RunComplete implements ReportNodeRun
func (r *PanelRun) RunComplete() bool {
	return r.runStatus == reportinterfaces.ReportRunComplete
}

// ChildrenComplete implements ReportNodeRun
func (r *PanelRun) ChildrenComplete() bool {
	return true
}
