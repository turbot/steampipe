package reportexecute

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ReportContainerRun is a struct representing a container run
type ReportContainerRun struct {
	Name          string                           `json:"name"`
	Title         string                           `json:"title,omitempty"`
	Width         int                              `json:"width,omitempty"`
	Height        int                              `json:"height,omitempty"`
	Source        string                           `json:"source,omitempty"`
	SQL           string                           `json:"sql,omitempty"`
	Error         error                            `json:"error,omitempty"`
	Children      []reportinterfaces.ReportNodeRun `json:"children,omitempty"`
	NodeType      string                           `json:"node_type"`
	Status        reportinterfaces.ReportRunStatus `json:"status"`
	Path          []string                         `json:"-"`
	parent        reportinterfaces.ReportNodeParent
	executionTree *ReportExecutionTree
	childComplete chan reportinterfaces.ReportNodeRun
}

func NewReportContainerRun(container *modconfig.ReportContainer, parent reportinterfaces.ReportNodeParent, executionTree *ReportExecutionTree) (*ReportContainerRun, error) {
	children := container.GetChildren()
	r := &ReportContainerRun{
		Name:          container.Name(),
		NodeType:      container.HclType,
		Path:          container.Paths[0],
		executionTree: executionTree,
		parent:        parent,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		Status:        reportinterfaces.ReportRunComplete,
		childComplete: make(chan reportinterfaces.ReportNodeRun, len(children)),
	}
	if container.Title != nil {
		r.Title = *container.Title
	}

	if container.Width != nil {
		r.Width = *container.Width
	}

	for _, child := range children {
		var childRun reportinterfaces.ReportNodeRun
		var err error
		switch i := child.(type) {
		case *modconfig.ReportContainer:
			childRun, err = NewReportContainerRun(i, r, executionTree)
			if err != nil {
				return nil, err
			}
		case *modconfig.Benchmark, *modconfig.Control:
			childRun, err = NewCheckRun(i.(modconfig.ReportingLeafNode), r, executionTree)
			if err != nil {
				return nil, err
			}

		default:
			// ensure this item is a ReportingLeafNode
			leafNode, ok := i.(modconfig.ReportingLeafNode)
			if !ok {
				return nil, fmt.Errorf("child %s does not implement ReportingLeafNode", i.Name())
			}

			childRun, err = NewLeafRun(leafNode, r, executionTree)
			if err != nil {
				return nil, err
			}
		}

		// should never happen - container children must be either container or counter
		if childRun == nil {
			continue
		}

		// if our child has not completed, we have not completed
		if childRun.GetRunStatus() == reportinterfaces.ReportRunReady {
			r.Status = reportinterfaces.ReportRunReady
		}
		r.Children = append(r.Children, childRun)
	}
	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r, nil
}

// Execute implements ReportRunNode
// execute all children and wait for them to complete
func (r *ReportContainerRun) Execute(ctx context.Context) error {

	errChan := make(chan error, len(r.Children))
	// execute all children asynchronously
	for _, child := range r.Children {
		go r.executeChild(ctx, child, errChan)
	}

	// wait for children to complete
	var errors []error
	for !r.ChildrenComplete() {
		select {
		case <-r.childComplete:
			// fall through to recheck ChildrenComplete
		case err := <-errChan:
			errors = append(errors, err)
			// TODO TIMEOUT??
		}
	}

	// so all children have completed - check for errors
	err := utils.CombineErrors(errors...)
	if err == nil {
		// set complete status on report - this will raise counter complete event
		r.SetComplete()
	} else {
		r.SetError(err)
	}

	return err
}

func (r *ReportContainerRun) executeChild(ctx context.Context, child reportinterfaces.ReportNodeRun, errChan chan error) {
	err := child.Execute(ctx)
	if err != nil {
		errChan <- err
	}
}

// GetName implements ReportNodeRun
func (r *ReportContainerRun) GetName() string {
	return r.Name
}

// GetPath implements ReportNodeRun
func (r *ReportContainerRun) GetPath() modconfig.NodePath {
	return r.Path
}

// GetRunStatus implements ReportNodeRun
func (r *ReportContainerRun) GetRunStatus() reportinterfaces.ReportRunStatus {
	return r.Status
}

// SetError implements ReportNodeRun
// tell parent we are done
func (r *ReportContainerRun) SetError(err error) {
	r.Error = err
	r.Status = reportinterfaces.ReportRunError
	// raise container error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ContainerError{Container: r})
	r.parent.ChildCompleteChan() <- r

}

// SetComplete implements ReportNodeRun
func (r *ReportContainerRun) SetComplete() {
	r.Status = reportinterfaces.ReportRunComplete
	// raise container complete event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ContainerComplete{Container: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements ReportNodeRun
func (r *ReportContainerRun) RunComplete() bool {
	return r.Status == reportinterfaces.ReportRunComplete || r.Status == reportinterfaces.ReportRunError
}

// ChildrenComplete implements ReportNodeRun
func (r *ReportContainerRun) ChildrenComplete() bool {
	for _, child := range r.Children {
		if !child.RunComplete() {
			return false
		}
	}

	return true
}

func (r *ReportContainerRun) ChildCompleteChan() chan reportinterfaces.ReportNodeRun {
	return r.childComplete
}
