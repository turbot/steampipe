package reportexecute

import (
	"context"

	"github.com/turbot/steampipe/control/controlexecute"
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
	Error error `json:"-"`
	// the parent report
	Report *modconfig.Report `json:"-"`

	// children
	PanelRuns  []*PanelRun
	ReportRuns []*ReportRun

	runStatus     ReportRunStatus
	executionTree *ReportExecutionTree
}

func NewReportRun(report *modconfig.Report, executionTree *controlexecute.ExecutionTree) *ReportRun {
	return &ReportRun{
		Report: report,
		// TODO OTHER STUFF
		runStatus: ReportRunReady,
	}
}

func (r *ReportRun) Start(ctx context.Context, client *db.Client) {

}

func (r *ReportRun) SetError(err error) {
	r.Error = err
	r.runStatus = ReportRunError
}
