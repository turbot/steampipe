package reportexecute

import (
	"fmt"

	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ReportContainerRun is a struct representing a container run
type ReportContainerRun struct {
	Name string `json:"name"`

	Text     string                           `json:"text,omitempty"`
	Type     string                           `json:"type,omitempty"`
	Width    int                              `json:"width,omitempty"`
	Height   int                              `json:"height,omitempty"`
	Source   string                           `json:"source,omitempty"`
	SQL      string                           `json:"sql,omitempty"`
	Data     [][]interface{}                  `json:"data,omitempty"`
	Error    error                            `json:"error,omitempty"`
	Children []reportinterfaces.ReportNodeRun `json:"children,omitempty"`

	runStatus     reportinterfaces.ReportRunStatus
	executionTree *ReportExecutionTree
}

func NewReportContainerRun(container *modconfig.ReportContainer, parentName string, executionTree *ReportExecutionTree) *ReportContainerRun {

	r := &ReportContainerRun{
		// the name is the path, i.e. dot-separated concatenation of parent names
		Name:          fmt.Sprintf("%s.%s", parentName, container.UnqualifiedName),
		executionTree: executionTree,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: reportinterfaces.ReportRunComplete,
	}
	if container.Width != nil {
		r.Width = *container.Width
	}

	for _, child := range container.GetChildren() {
		var childRun reportinterfaces.ReportNodeRun
		switch i := child.(type) {
		case *modconfig.ReportContainer:
			childRun = NewReportContainerRun(i, r.Name, executionTree)
		case *modconfig.Panel:
			childRun = NewPanelRun(i, r.Name, executionTree)
		}

		// should never happen - container children must be either container or panel
		if childRun == nil {
			continue
		}

		// if our child has not completed, we have not completed
		if childRun.GetRunStatus() == reportinterfaces.ReportRunReady {
			// add dependency on this child
			r.executionTree.AddDependency(r.Name, childRun.GetName())
			r.runStatus = reportinterfaces.ReportRunReady
		}
		r.Children = append(r.Children, childRun)
	}
	// add r into execution tree
	if container.IsReport() {
		executionTree.reports[r.Name] = r
	} else {
		executionTree.containers[r.Name] = r
	}
	return r
}

// GetName implements ReportNodeRun
func (r *ReportContainerRun) GetName() string {
	return r.Name
}

// GetRunStatus implements ReportNodeRun
func (r *ReportContainerRun) GetRunStatus() reportinterfaces.ReportRunStatus {
	return r.runStatus
}

// SetError implements ReportNodeRun
func (r *ReportContainerRun) SetError(err error) {
	r.Error = err
	r.runStatus = reportinterfaces.ReportRunError
	// raise container error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ContainerError{Container: r})

}

// SetComplete implements ReportNodeRun
func (r *ReportContainerRun) SetComplete() {
	r.runStatus = reportinterfaces.ReportRunComplete
	// raise container complete event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ContainerComplete{Container: r})
}

// RunComplete implements ReportNodeRun
func (r *ReportContainerRun) RunComplete() bool {
	return r.runStatus == reportinterfaces.ReportRunComplete
}

// ChildrenComplete implements ReportNodeRun
func (r *ReportContainerRun) ChildrenComplete() bool {
	for _, container := range r.Children {
		if container.GetRunStatus() != reportinterfaces.ReportRunComplete {
			return false
		}
	}
	return true
}
