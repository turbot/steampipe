package dashboardexecute

import (
	"context"
	"fmt"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// DashboardContainerRun is a struct representing a container run
type DashboardContainerRun struct {
	Name             string                            `json:"name"`
	Title            string                            `json:"title,omitempty"`
	Width            int                               `json:"width,omitempty"`
	Display          string                            `json:"display,omitempty"`
	ErrorString      string                            `json:"error,omitempty"`
	NodeType         string                            `json:"panel_type"`
	Status           dashboardtypes.DashboardRunStatus `json:"status"`
	DashboardName    string                            `json:"dashboard"`
	SourceDefinition string                            `json:"source_definition"`

	children      []dashboardtypes.DashboardNodeRun
	error         error
	dashboardNode *modconfig.DashboardContainer
	parent        dashboardtypes.DashboardNodeParent
	executionTree *DashboardExecutionTree
	childComplete chan dashboardtypes.DashboardNodeRun
}

func (r *DashboardContainerRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	res := &dashboardtypes.SnapshotTreeNode{
		Name:     r.Name,
		NodeType: r.NodeType,
		Children: make([]*dashboardtypes.SnapshotTreeNode, len(r.children)),
	}
	for i, c := range r.children {
		res.Children[i] = c.AsTreeNode()
	}
	return res
}

func NewDashboardContainerRun(container *modconfig.DashboardContainer, parent dashboardtypes.DashboardNodeParent, executionTree *DashboardExecutionTree) (*DashboardContainerRun, error) {
	children := container.GetChildren()

	// NOTE: for now we MUST declare children inline - therefore we cannot share children between runs in the tree
	// (if we supported the children property then we could reuse resources)
	// so FOR NOW it is safe to use the container name directly as the run name
	name := container.Name()

	r := &DashboardContainerRun{
		Name:             name,
		NodeType:         modconfig.BlockTypeContainer,
		DashboardName:    executionTree.dashboardName,
		Display:          typehelpers.SafeString(container.Display),
		SourceDefinition: container.GetMetadata().SourceDefinition,
		executionTree:    executionTree,
		parent:           parent,
		dashboardNode:    container,

		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		Status:        dashboardtypes.DashboardRunComplete,
		childComplete: make(chan dashboardtypes.DashboardNodeRun, len(children)),
	}
	if container.Title != nil {
		r.Title = *container.Title
	}

	if container.Width != nil {
		r.Width = *container.Width
	}

	for _, child := range children {
		var childRun dashboardtypes.DashboardNodeRun
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
		if childRun.GetRunStatus() == dashboardtypes.DashboardRunReady {
			r.Status = dashboardtypes.DashboardRunReady
		}
		r.children = append(r.children, childRun)
	}
	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r, nil
}

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardContainerRun) IsSnapshotPanel() {}

// Initialise implements DashboardRunNode
func (r *DashboardContainerRun) Initialise(ctx context.Context) {
	// initialise our children
	for _, child := range r.children {
		child.Initialise(ctx)
		if err := child.GetError(); err != nil {
			r.SetError(ctx, err)
			return
		}
	}
}

// Execute implements DashboardRunNode
// execute all children and wait for them to complete
func (r *DashboardContainerRun) Execute(ctx context.Context) {
	// execute all children asynchronously
	for _, child := range r.children {
		go child.Execute(ctx)
	}

	// wait for children to complete
	var errors []error
	for !r.ChildrenComplete() {
		completeChild := <-r.childComplete
		if completeChild.GetRunStatus() == dashboardtypes.DashboardRunError {
			errors = append(errors, completeChild.GetError())
		}
		// fall through to recheck ChildrenComplete
		// TODO [reports]  timeout?
	}

	// so all children have completed - check for errors
	err := error_helpers.CombineErrors(errors...)
	if err == nil {
		// set complete status on dashboard
		r.SetComplete(ctx)
	} else {
		r.SetError(ctx, err)
	}
}

// GetName implements DashboardNodeRun
func (r *DashboardContainerRun) GetName() string {
	return r.Name
}

// GetRunStatus implements DashboardNodeRun
func (r *DashboardContainerRun) GetRunStatus() dashboardtypes.DashboardRunStatus {
	return r.Status
}

// SetError implements DashboardNodeRun
// tell parent we are done
func (r *DashboardContainerRun) SetError(_ context.Context, err error) {
	r.error = err
	// error type does not serialise to JSON so copy into a string
	r.ErrorString = err.Error()
	r.Status = dashboardtypes.DashboardRunError
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

// SetComplete implements DashboardNodeRun
func (r *DashboardContainerRun) SetComplete(context.Context) {
	r.Status = dashboardtypes.DashboardRunComplete
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
	return r.Status == dashboardtypes.DashboardRunComplete || r.Status == dashboardtypes.DashboardRunError
}

// GetChildren implements DashboardNodeRun
func (r *DashboardContainerRun) GetChildren() []dashboardtypes.DashboardNodeRun {
	return r.children
}

// ChildrenComplete implements DashboardNodeRun
func (r *DashboardContainerRun) ChildrenComplete() bool {
	for _, child := range r.children {
		if !child.RunComplete() {
			return false
		}
	}

	return true
}

// GetTitle implements DashboardNodeRun
func (r *DashboardContainerRun) GetTitle() string {
	return r.Title
}

// ChildCompleteChan implements DashboardNodeParent
func (r *DashboardContainerRun) ChildCompleteChan() chan dashboardtypes.DashboardNodeRun {
	return r.childComplete
}

// GetInputsDependingOn implements DashboardNodeRun
// return nothing for DashboardContainerRun
func (r *DashboardContainerRun) GetInputsDependingOn(changedInputName string) []string { return nil }
