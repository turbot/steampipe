package reportexecute

import (
	"context"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type PanelRunStatus uint32

const (
	PanelRunReady PanelRunStatus = 1 << iota
	PanelRunStarted
	PanelRunComplete
	PanelRunError
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

	runStatus     PanelRunStatus
	executionTree *ReportExecutionTree
}

func NewPanelRun(panel *modconfig.Panel, executionTree *ReportExecutionTree) *PanelRun {
	r := &PanelRun{
		Name:          panel.Name(),
		Title:         typehelpers.SafeString(panel.Title),
		Text:          typehelpers.SafeString(panel.Text),
		Source:        typehelpers.SafeString(panel.Source),
		SQL:           typehelpers.SafeString(panel.SQL),
		executionTree: executionTree,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: PanelRunComplete,
	}
	if panel.Width != nil {
		r.Width = *panel.Width
	}

	// if we have sql, set status to ready
	if panel.SQL != nil {
		r.runStatus = PanelRunReady
	}
	// create report runs for all children
	for _, childReport := range panel.Reports {
		// todo register dependencies
		childRun := NewReportRun(childReport, executionTree)
		// if our child has not completed, we have not completed
		if childRun.runStatus == ReportRunReady {
			// add a dependency on this child
			executionTree.AddDependency(r.Name, childRun.Name)
			r.runStatus = PanelRunReady
		}
		r.ReportRuns = append(r.ReportRuns, childRun)
	}
	for _, childPanel := range panel.Panels {
		// todo register dependencies
		childRun := NewPanelRun(childPanel, executionTree)
		// if our child has not completed, we have not completed
		if childRun.runStatus == PanelRunReady {
			r.runStatus = PanelRunReady
		}
		r.PanelRuns = append(r.PanelRuns, childRun)
	}
	return r
}

func (r *PanelRun) Start(ctx context.Context, client *db.Client) {

}

func (r *PanelRun) SetError(err error) {
	r.Error = err
	r.runStatus = PanelRunError
}
