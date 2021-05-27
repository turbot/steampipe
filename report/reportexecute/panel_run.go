package reportexecute

import (
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportexecutiontree"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// PanelRun is a struct representing a  a panel run - will contain one or more result items (i.e. for one or more resources)
type PanelRun struct {
	Name   string          `json:"name"`
	Title  string          `json:"title,omitempty"`
	Text   string          `json:"text,omitempty"`
	Width  int             `json:"width,omitempty"`
	Source string          `json:"source,omitempty"`
	SQL    string          `json:"sql,omitempty"`
	Data   [][]interface{} `json:"data,omitempty"`

	Error error `json:"-"`

	// children
	PanelRuns  []*PanelRun  `json:"panels,omitempty"`
	ReportRuns []*ReportRun `json:"reports,omitempty"`

	runStatus     ReportRunStatus
	executionTree *reportexecutiontree.ReportExecutionTree
}

func NewPanelRun(panel *modconfig.Panel, executionTree *reportexecutiontree.ReportExecutionTree) *PanelRun {
	r := &PanelRun{
		Name:          panel.Name(),
		Title:         typehelpers.SafeString(panel.Title),
		Text:          typehelpers.SafeString(panel.Text),
		Source:        typehelpers.SafeString(panel.Source),
		SQL:           typehelpers.SafeString(panel.SQL),
		executionTree: executionTree,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: ReportRunComplete,
	}
	if panel.Width != nil {
		r.Width = *panel.Width
	}

	// if we have sql, set status to ready
	if panel.SQL != nil {
		r.runStatus = ReportRunReady
	}
	// create report runs for all children
	for _, childReport := range panel.Reports {
		childRun := NewReportRun(childReport, executionTree)
		// if our child has not completed, we have not completed
		if childRun.runStatus == ReportRunReady {
			// add a dependency on this child
			executionTree.AddDependency(r.Name, childRun.Name)
			r.runStatus = ReportRunReady
		}
		r.ReportRuns = append(r.ReportRuns, childRun)
	}
	for _, childPanel := range panel.Panels {
		childRun := NewPanelRun(childPanel, executionTree)
		// if our child has not completed, we have not completed
		if childRun.runStatus == ReportRunReady {
			// add dependency on this child
			r.executionTree.AddDependency(r.Name, childRun.Name)
			r.runStatus = ReportRunReady
		}
		r.PanelRuns = append(r.PanelRuns, childRun)
	}
	// add r into execution tree
	executionTree.panels[r.Name] = r
	return r
}

func (r *PanelRun) SetError(err error) {
	r.Error = err
	r.runStatus = ReportRunError
}

func (r *PanelRun) SetComplete() {
	r.runStatus = ReportRunComplete
	// raise panel complete event
	r.executionTree.workspace.PublishReportEvent(&reportevents.PanelComplete{Panel: r})
}

func (r *PanelRun) ChildrenComplete() bool {
	for _, panel := range r.PanelRuns {
		if panel.runStatus != ReportRunComplete {
			return false
		}
	}
	for _, report := range r.ReportRuns {
		if report.runStatus != ReportRunComplete {
			return false
		}
	}
	return true
}
