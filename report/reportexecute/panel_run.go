package reportexecute

import (
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// PanelRun is a struct representing a  a panel run - will contain one or more result items (i.e. for one or more resources)
type PanelRun struct {
	Name   string          `json:"name"`
	Title  string          `json:"title,omitempty"`
	Text   string          `json:"text,omitempty"`
	Type   string          `json:"type,omitempty"`
	Width  int             `json:"width,omitempty"`
	Height int             `json:"height,omitempty"`
	Source string          `json:"source,omitempty"`
	SQL    string          `json:"sql,omitempty"`
	Data   [][]interface{} `json:"data,omitempty"`

	Error error `json:"error,omitempty"`

	// children
	PanelRuns  []*PanelRun  `json:"panels,omitempty"`
	ReportRuns []*ReportRun `json:"reports,omitempty"`

	runStatus     reportinterfaces.ReportRunStatus
	executionTree *ReportExecutionTree
}

func NewPanelRun(panel *modconfig.Panel, executionTree *ReportExecutionTree) *PanelRun {
	r := &PanelRun{
		Name:          panel.Name(),
		Title:         typehelpers.SafeString(panel.Title),
		Text:          typehelpers.SafeString(panel.Text),
		Type:          typehelpers.SafeString(panel.Type),
		Source:        typehelpers.SafeString(panel.Source),
		SQL:           typehelpers.SafeString(panel.SQL),
		executionTree: executionTree,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: reportinterfaces.ReportRunComplete,
	}
	if panel.Width != nil {
		r.Width = *panel.Width
	}

	if panel.Height != nil {
		r.Height = *panel.Height
	}

	// if we have sql, set status to ready
	if panel.SQL != nil {
		r.runStatus = reportinterfaces.ReportRunReady
	}

	for _, childPanel := range panel.Panels {
		childRun := NewPanelRun(childPanel, executionTree)
		// if our child has not completed, we have not completed
		if childRun.runStatus == reportinterfaces.ReportRunReady {
			// add dependency on this child
			r.executionTree.AddDependency(r.Name, childRun.Name)
			r.runStatus = reportinterfaces.ReportRunReady
		}
		r.PanelRuns = append(r.PanelRuns, childRun)
	}
	// add r into execution tree
	executionTree.panels[r.Name] = r
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
	for _, panel := range r.PanelRuns {
		if panel.runStatus != reportinterfaces.ReportRunComplete {
			return false
		}
	}
	for _, report := range r.ReportRuns {
		if report.runStatus != reportinterfaces.ReportRunComplete {
			return false
		}
	}
	return true
}
