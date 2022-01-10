package reportexecute

import (
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ContainerRun is a struct representing a container run
type ContainerRun struct {
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
	ContainerRuns []*ContainerRun `json:"containers,omitempty"`
	ReportRuns    []*ReportRun    `json:"reports,omitempty"`

	runStatus     reportinterfaces.ReportRunStatus
	executionTree *ReportExecutionTree
}

func NewContainerRun(container *modconfig.Container, executionTree *ReportExecutionTree) *ContainerRun {
	r := &ContainerRun{
		Name:          container.Name(),
		executionTree: executionTree,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: reportinterfaces.ReportRunComplete,
	}
	if container.Width != nil {
		r.Width = *container.Width
	}

	for _, childContainer := range container.Containers {
		childRun := NewContainerRun(childContainer, executionTree)
		// if our child has not completed, we have not completed
		if childRun.runStatus == reportinterfaces.ReportRunReady {
			// add dependency on this child
			r.executionTree.AddDependency(r.Name, childRun.Name)
			r.runStatus = reportinterfaces.ReportRunReady
		}
		r.ContainerRuns = append(r.ContainerRuns, childRun)
	}
	// add r into execution tree
	executionTree.containers[r.Name] = r
	return r
}

// GetName implements ReportNodeRun
func (r *ContainerRun) GetName() string {
	return r.Name
}

// GetRunStatus implements ReportNodeRun
func (r *ContainerRun) GetRunStatus() reportinterfaces.ReportRunStatus {
	return r.runStatus
}

// SetError implements ReportNodeRun
func (r *ContainerRun) SetError(err error) {
	r.Error = err
	r.runStatus = reportinterfaces.ReportRunError
	// raise container error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ContainerError{Container: r})

}

// SetComplete implements ReportNodeRun
func (r *ContainerRun) SetComplete() {
	r.runStatus = reportinterfaces.ReportRunComplete
	// raise container complete event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ContainerComplete{Container: r})
}

// RunComplete implements ReportNodeRun
func (r *ContainerRun) RunComplete() bool {
	return r.runStatus == reportinterfaces.ReportRunComplete
}

// ChildrenComplete implements ReportNodeRun
func (r *ContainerRun) ChildrenComplete() bool {
	for _, container := range r.ContainerRuns {
		if container.runStatus != reportinterfaces.ReportRunComplete {
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
