package reportexecute

import (
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportexecutiontree"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ReportRunStatus uint32

const (
	ReportRunReady ReportRunStatus = 1 << iota
	ReportRunStarted
	ReportRunComplete
	ReportRunError
)

// ReportRun is a struct representing a  a report run - will contain one or more result items (i.e. for one or more resources)
type ReportRun struct {
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`

	// children
	PanelRuns  []*PanelRun  `json:"panels,omitempty"`
	ReportRuns []*ReportRun `json:"reports,omitempty"`

	Error error `json:"-"`

	runStatus     ReportRunStatus                          `json:"-"`
	executionTree *reportexecutiontree.ReportExecutionTree `json:"-"`
}

func NewReportRun(report *modconfig.Report, executionTree *reportexecutiontree.ReportExecutionTree) *ReportRun {
	r := &ReportRun{
		Name:          report.Name(),
		Title:         typehelpers.SafeString(report.Title),
		executionTree: executionTree,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: ReportRunComplete,
	}

	// create report runs for all children
	for _, childReport := range report.Reports {
		childRun := NewReportRun(childReport, executionTree)
		// if our child has not completed, we have not completed
		if childRun.runStatus == ReportRunReady {
			// add dependency on this child
			r.executionTree.AddDependency(r.Name, childRun.Name)
			r.runStatus = ReportRunReady
		}
		r.ReportRuns = append(r.ReportRuns, childRun)
	}
	for _, childPanel := range report.Panels {
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
	executionTree.reports[r.Name] = r
	return r
}

func (r *ReportRun) SetError(err error) {
	r.Error = err
	r.runStatus = ReportRunError
	// raise report error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ReportError{Report: r})
}

func (r *ReportRun) SetComplete() {
	r.runStatus = ReportRunComplete
	// raise report complete event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ReportComplete{Report: r})
}

func (r *ReportRun) ChildrenComplete() bool {
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
