package reportexecute

import (
	"context"

	typehelpers "github.com/turbot/go-kit/types"

	"github.com/turbot/steampipe/db"
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

	runStatus     ReportRunStatus      `json:"-"`
	executionTree *ReportExecutionTree `json:"-"`
}

func NewReportRun(report *modconfig.Report, executionTree *ReportExecutionTree) *ReportRun {
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
		// todo register dependencies
		childRun := NewReportRun(childReport, executionTree)
		// if our child has not completed, we have not completed
		if childRun.runStatus == ReportRunReady {
			r.runStatus = ReportRunReady
		}
		r.ReportRuns = append(r.ReportRuns, childRun)
	}
	return r
}

func (r *ReportRun) Start(ctx context.Context, client *db.Client) {

}

func (r *ReportRun) SetError(err error) {
	r.Error = err
	r.runStatus = ReportRunError
}
