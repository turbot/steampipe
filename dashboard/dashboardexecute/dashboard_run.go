package dashboardexecute

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// TODO [reports] split into report and container
// update events
// DashboardRun is a struct representing a container run

type DashboardRun struct {
	Name          string                                 `json:"name"`
	Title         string                                 `json:"title,omitempty"`
	Width         int                                    `json:"width,omitempty"`
	Height        int                                    `json:"height,omitempty"`
	Source        string                                 `json:"source,omitempty"`
	SQL           string                                 `json:"sql,omitempty"`
	Error         error                                  `json:"error,omitempty"`
	Children      []dashboardinterfaces.DashboardNodeRun `json:"children,omitempty"`
	NodeType      string                                 `json:"node_type"`
	Status        dashboardinterfaces.DashboardRunStatus `json:"status"`
	DashboardName string                                 `json:"dashboard"`
	dashboardNode *modconfig.Dashboard
	parent        dashboardinterfaces.DashboardNodeParent
	executionTree *DashboardExecutionTree
	childComplete chan dashboardinterfaces.DashboardNodeRun
}

func NewDashboardRun(dashboard *modconfig.Dashboard, parent dashboardinterfaces.DashboardNodeParent, executionTree *DashboardExecutionTree) (*DashboardRun, error) {
	children := dashboard.GetChildren()

	// NOTE: for now we MUST declare container/dashboard children inline - therefore we cannot share children between runs in the tree
	// (if we supported the children property then we could reuse resources)
	// so FOR NOW it is safe to use the dashboard name directly as the run name
	name := dashboard.Name()

	r := &DashboardRun{
		Name:          name,
		NodeType:      modconfig.BlockTypeDashboard,
		DashboardName: executionTree.dashboardName,
		executionTree: executionTree,
		parent:        parent,
		dashboardNode: dashboard,

		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		Status:        dashboardinterfaces.DashboardRunComplete,
		childComplete: make(chan dashboardinterfaces.DashboardNodeRun, len(children)),
	}
	if dashboard.Title != nil {
		r.Title = *dashboard.Title
	}

	if dashboard.Width != nil {
		r.Width = *dashboard.Width
	}

	for _, child := range children {
		var childRun dashboardinterfaces.DashboardNodeRun
		var err error
		switch i := child.(type) {
		case *modconfig.Dashboard:
			childRun, err = NewDashboardRun(i, r, executionTree)
			if err != nil {
				return nil, err
			}
		case *modconfig.DashboardContainer:
			childRun, err = NewDashboardContainerRun(i, r, executionTree)
			if err != nil {
				return nil, err
			}
		case *modconfig.Benchmark, *modconfig.Control:
			childRun, err = NewCheckRun(i.(modconfig.DashboardLeafNode), r, executionTree)
			if err != nil {
				return nil, err
			}
		case *modconfig.DashboardInput:
			// NOTE:L clone the input to avoid mutating the original
			childRun, err = NewLeafRun(i.Clone(), r, executionTree)
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
func (r *DashboardRun) Execute(ctx context.Context) error {
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
		// set complete status on dashboard
		r.SetComplete()
	} else {
		r.SetError(err)
	}

	return err
}

func (r *DashboardRun) executeChild(ctx context.Context, child dashboardinterfaces.DashboardNodeRun, errChan chan error) {
	err := child.Execute(ctx)
	if err != nil {
		errChan <- err
	}
}

// GetName implements DashboardNodeRun
func (r *DashboardRun) GetName() string {
	return r.Name
}

// GetRunStatus implements DashboardNodeRun
func (r *DashboardRun) GetRunStatus() dashboardinterfaces.DashboardRunStatus {
	return r.Status
}

// SetError implements DashboardNodeRun
// tell parent we are done
func (r *DashboardRun) SetError(err error) {
	r.Error = err
	r.Status = dashboardinterfaces.DashboardRunError
	// raise container error event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.ContainerError{
		Container: r,
		Session:   r.executionTree.sessionId,
	})
	r.parent.ChildCompleteChan() <- r

}

// SetComplete implements DashboardNodeRun
func (r *DashboardRun) SetComplete() {
	r.Status = dashboardinterfaces.DashboardRunComplete
	// raise container complete event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.ContainerComplete{
		Container: r,
		Session:   r.executionTree.sessionId,
	})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements DashboardNodeRun
func (r *DashboardRun) RunComplete() bool {
	return r.Status == dashboardinterfaces.DashboardRunComplete || r.Status == dashboardinterfaces.DashboardRunError
}

// ChildrenComplete implements DashboardNodeRun
func (r *DashboardRun) ChildrenComplete() bool {
	for _, child := range r.Children {
		if !child.RunComplete() {
			return false
		}
	}

	return true
}

func (r *DashboardRun) ChildCompleteChan() chan dashboardinterfaces.DashboardNodeRun {
	return r.childComplete
}

// GetInput searches for an input with the given name
func (r *DashboardRun) GetInput(name string) (*modconfig.DashboardInput, bool) {
	return r.dashboardNode.GetInput(name)
}
