package dashboardexecute

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// DashboardContainerRun is a struct representing a container run
type DashboardContainerRun struct {
	Name             string                                 `json:"name"`
	Title            string                                 `json:"title,omitempty"`
	Width            int                                    `json:"width,omitempty"`
	ErrorString      string                                 `json:"error,omitempty"`
	Children         []dashboardinterfaces.DashboardNodeRun `json:"children,omitempty"`
	NodeType         string                                 `json:"node_type"`
	Status           dashboardinterfaces.DashboardRunStatus `json:"status"`
	DashboardName    string                                 `json:"report"`
	SourceDefinition string                                 `json:"source_definition"`
	error            error
	dashboardNode    *modconfig.DashboardContainer
	parent           dashboardinterfaces.DashboardNodeParent
	executionTree    *DashboardExecutionTree
	childComplete    chan dashboardinterfaces.DashboardNodeRun
}

func NewDashboardContainerRun(container *modconfig.DashboardContainer, parent dashboardinterfaces.DashboardNodeParent, executionTree *DashboardExecutionTree) (*DashboardContainerRun, error) {
	children := container.GetChildren()

	// NOTE: for now we MUST declare children inline - therefore we cannot share children between runs in the tree
	// (if we supported the children property then we could reuse resources)
	// so FOR NOW it is safe to use the container name directly as the run name
	name := container.Name()

	r := &DashboardContainerRun{
		Name:             name,
		NodeType:         modconfig.BlockTypeContainer,
		DashboardName:    executionTree.dashboardName,
		SourceDefinition: container.GetMetadata().SourceDefinition,
		executionTree:    executionTree,
		parent:           parent,
		dashboardNode:    container,

		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		Status:        dashboardinterfaces.DashboardRunComplete,
		childComplete: make(chan dashboardinterfaces.DashboardNodeRun, len(children)),
	}
	if container.Title != nil {
		r.Title = *container.Title
	}

	if container.Width != nil {
		r.Width = *container.Width
	}

	for _, child := range children {
		var childRun dashboardinterfaces.DashboardNodeRun
		var err error
		switch i := child.(type) {
		case *modconfig.DashboardContainer:
			childRun, err = NewDashboardContainerRun(i, r, executionTree)
			if err != nil {
				return nil, err
			}
		case *modconfig.Dashboard:
			childRun, err = NewDashboardRun(i, r, executionTree)
			if err != nil {
				return nil, err
			}
		case *modconfig.Benchmark, *modconfig.Control:
			childRun, err = NewCheckRun(i.(modconfig.DashboardLeafNode), r, executionTree)
			if err != nil {
				return nil, err
			}

		default:
			// ensure this item is a DashboardLeafNode
			leafNode, ok := i.(modconfig.DashboardLeafNode)
			if !ok {
				return nil, fmt.Errorf("child %s does not implement DashboardLeafNode", i.Name())
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
		if childRun.GetRunStatus() == dashboardinterfaces.DashboardRunReady {
			r.Status = dashboardinterfaces.DashboardRunReady
		}
		r.Children = append(r.Children, childRun)
	}
	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r, nil
}

// Execute implements DashboardRunNode
// execute all children and wait for them to complete
func (r *DashboardContainerRun) Execute(ctx context.Context) {
	// execute all children asynchronously
	for _, child := range r.Children {
		go child.Execute(ctx)
	}

	// wait for children to complete
	var errors []error
	for !r.ChildrenComplete() {
		select {
		case completeChild := <-r.childComplete:
			if completeChild.GetRunStatus() == dashboardinterfaces.DashboardRunError {
				errors = append(errors, completeChild.GetError())
			}
			// fall through to recheck ChildrenComplete

			// TODO [reports]  timeout?
		}
	}

	// so all children have completed - check for errors
	err := utils.CombineErrors(errors...)
	if err == nil {
		// set complete status on dashboard
		r.SetComplete()
	} else {
		r.SetError(err)
	}
}

// GetName implements DashboardNodeRun
func (r *DashboardContainerRun) GetName() string {
	return r.Name
}

// GetRunStatus implements DashboardNodeRun
func (r *DashboardContainerRun) GetRunStatus() dashboardinterfaces.DashboardRunStatus {
	return r.Status
}

// SetError implements DashboardNodeRun
// tell parent we are done
func (r *DashboardContainerRun) SetError(err error) {
	r.error = err
	// error type does not serialise to JSON so copy into a string
	r.ErrorString = err.Error()
	r.Status = dashboardinterfaces.DashboardRunError
	// raise container error event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.ContainerError{
		Container:   r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
	})
	r.parent.ChildCompleteChan() <- r
}

// GetError implements DashboardNodeRun
func (r *DashboardContainerRun) GetError() error {
	return r.error
}

// GetChildren implements DashboardNodeRun
func (r *DashboardContainerRun) GetChildren() []dashboardinterfaces.DashboardNodeRun {
	return r.Children
}

// SetComplete implements DashboardNodeRun
func (r *DashboardContainerRun) SetComplete() {
	r.Status = dashboardinterfaces.DashboardRunComplete
	// raise container complete event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.ContainerComplete{
		Container:   r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
	})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements DashboardNodeRun
func (r *DashboardContainerRun) RunComplete() bool {
	return r.Status == dashboardinterfaces.DashboardRunComplete || r.Status == dashboardinterfaces.DashboardRunError
}

// ChildrenComplete implements DashboardNodeRun
func (r *DashboardContainerRun) ChildrenComplete() bool {
	for _, child := range r.Children {
		if !child.RunComplete() {
			return false
		}
	}

	return true
}

func (r *DashboardContainerRun) ChildCompleteChan() chan dashboardinterfaces.DashboardNodeRun {
	return r.childComplete
}

// GetInputsDependingOn implements DashboardNodeRun
//return nothing for DashboardContainerRun
func (r *DashboardContainerRun) GetInputsDependingOn(changedInputName string) []string { return nil }
