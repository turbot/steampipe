package reportexecute

import (
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ReportRun is a struct representing a  a report run - will contain one or more result items (i.e. for one or more resources)
type ReportRun struct {
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`

	// children
	PanelRuns     []*PanelRun     `json:"panels,omitempty"`
	ContainerRuns []*ContainerRun `json:"containers,omitempty"`

	Error error `json:"error,omitempty"`

	runStatus     reportinterfaces.ReportRunStatus `json:"-"`
	executionTree *ReportExecutionTree             `json:"-"`
}

func NewReportRun(report *modconfig.Report, executionTree *ReportExecutionTree) *ReportRun {
	r := &ReportRun{
		Name:          report.Name(),
		Title:         typehelpers.SafeString(report.Title),
		executionTree: executionTree,

		// set to complete, optimistically
		// if any children have SQL we will set this to reportinterfaces.ReportRunReady instead
		runStatus: reportinterfaces.ReportRunComplete,
	}

	// create container runs for all children
	for _, childContainer := range report.Containers {
		childRun := NewContainerRun(childContainer, executionTree)
		// if our child has not completed, we have not completed
		if childRun.runStatus == reportinterfaces.ReportRunReady {
			// add dependency on this child
			r.executionTree.AddDependency(r.Name, childRun.Name)
			r.runStatus = reportinterfaces.ReportRunReady
		}
		r.ContainerRuns = append(r.ContainerRuns, childRun)
	}
	for _, childPanel := range report.Panels {
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
	executionTree.reports[r.Name] = r
	return r
}

// GetName implements ReportNodeRun
func (r *ReportRun) GetName() string {
	return r.Name
}

// GetRunStatus implements ReportNodeRun
func (r *ReportRun) GetRunStatus() reportinterfaces.ReportRunStatus {
	return r.runStatus
}

// SetError implements ReportNodeRun
func (r *ReportRun) SetError(err error) {
	r.Error = err
	r.runStatus = reportinterfaces.ReportRunError
	// raise report error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ReportError{Report: r})
}

// SetComplete implements ReportNodeRun
func (r *ReportRun) SetComplete() {
	r.runStatus = reportinterfaces.ReportRunComplete
	// raise report complete event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ReportComplete{Report: r})
}

// RunComplete implements ReportNodeRun
func (r *ReportRun) RunComplete() bool {
	return r.runStatus == reportinterfaces.ReportRunComplete
}

// ChildrenComplete implements ReportNodeRun
func (r *ReportRun) ChildrenComplete() bool {
	for _, panel := range r.PanelRuns {
		if !panel.RunComplete() {
			return false
		}
	}
	for _, report := range r.ContainerRuns {
		if !report.RunComplete() {
			return false
		}
	}
	return true
}
