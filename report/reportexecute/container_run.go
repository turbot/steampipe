package reportexecute

import (
	"context"
	"fmt"
	"log"

	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ReportContainerRun is a struct representing a container run
type ReportContainerRun struct {
	Name string `json:"name"`

	Width    int                              `json:"width,omitempty"`
	Height   int                              `json:"height,omitempty"`
	Source   string                           `json:"source,omitempty"`
	SQL      string                           `json:"sql,omitempty"`
	Error    error                            `json:"error,omitempty"`
	Children []reportinterfaces.ReportNodeRun `json:"children,omitempty"`

	parent        reportinterfaces.ReportNodeParent
	runStatus     reportinterfaces.ReportRunStatus
	executionTree *ReportExecutionTree
	childComplete chan (reportinterfaces.ReportNodeRun)
}

func NewReportContainerRun(container *modconfig.ReportContainer, parent reportinterfaces.ReportNodeParent, executionTree *ReportExecutionTree) (*ReportContainerRun, error) {

	children := container.GetChildren()
	r := &ReportContainerRun{
		Name:          container.Name(),
		executionTree: executionTree,
		parent:        parent,
		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus:     reportinterfaces.ReportRunComplete,
		childComplete: make(chan reportinterfaces.ReportNodeRun, len(children)),
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
			r.runStatus = reportinterfaces.ReportRunReady
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
	log.Printf("[WARN] %s Execute", r.Name)

	errChan := make(chan error, len(r.Children))
	// execute all children asynchronously
	for _, child := range r.Children {
		go r.executeChild(ctx, child, errChan)
	}

	//log.Printf("[WARN] %s wait for completion", r.Name)
	// wait for children to complete
	var errors []error
	for !r.ChildrenComplete() {
		select {
		case <-r.childComplete:
			// recheck ChildrenComplete
			//log.Printf("[WARN] %s got childComplete from %s", r.Name, child.GetName())

		case err := <-errChan:
			errors = append(errors, err)
			// TODO TIMEOUT??
		}
	}

	//log.Printf("[WARN] %s ChildrenComplete", r.Name)

	// so all children have completed - check for errors
	err := utils.CombineErrors(errors...)
	if err == nil {
		log.Printf("[WARN] %s ALL DONE", r.Name)
		// set complete status on report - this will raise counter complete event
		r.SetComplete()
	} else {
		r.SetError(err)
	}

	return err
}

func (r *ReportContainerRun) executeChild(ctx context.Context, child reportinterfaces.ReportNodeRun, errChan chan error) {
	//log.Printf("[WARN] %s call Execute for %s", r.Name, child.GetName())

	err := child.Execute(ctx)
	if err != nil {
		errChan <- err
	}
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
// tell parent we are done
func (r *ReportContainerRun) SetError(err error) {
	r.Error = err
	r.runStatus = reportinterfaces.ReportRunError
	// raise container error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ContainerError{Container: r})
	r.parent.ChildCompleteChan() <- r

}

// SetComplete implements ReportNodeRun
func (r *ReportContainerRun) SetComplete() {
	r.runStatus = reportinterfaces.ReportRunComplete
	// raise container complete event
	r.executionTree.workspace.PublishReportEvent(&reportevents.ContainerComplete{Container: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements ReportNodeRun
func (r *ReportContainerRun) RunComplete() bool {
	return r.runStatus == reportinterfaces.ReportRunComplete || r.runStatus == reportinterfaces.ReportRunError
}

// ChildrenComplete implements ReportNodeRun
func (r *ReportContainerRun) ChildrenComplete() bool {
	//log.Printf("[WARN] %s ChildrenComplete", r.Name)
	for _, child := range r.Children {
		if !child.RunComplete() {
			log.Printf("[WARN] %s child %s is not complete", r.Name, child.GetName())
			return false
		}
	}

	return true
}

func (r *ReportContainerRun) ChildCompleteChan() chan reportinterfaces.ReportNodeRun {
	return r.childComplete
}
