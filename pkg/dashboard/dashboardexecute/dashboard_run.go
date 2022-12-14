package dashboardexecute

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"log"
)

// DashboardRun is a struct representing a container run
type DashboardRun struct {
	RuntimeDependencyPublisherImpl

	parent    dashboardtypes.DashboardParent
	dashboard *modconfig.Dashboard
}

func (r *DashboardRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	res := &dashboardtypes.SnapshotTreeNode{
		Name:     r.Name,
		NodeType: r.NodeType,
		Children: make([]*dashboardtypes.SnapshotTreeNode, 0, len(r.children)),
	}

	for _, c := range r.children {
		// NOTE: exclude with runs
		if c.GetNodeType() != modconfig.BlockTypeWith {
			res.Children = append(res.Children, c.AsTreeNode())
		}
	}

	return res
}

// TODO can dashboards have params????
func NewDashboardRun(dashboard *modconfig.Dashboard, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) (*DashboardRun, error) {
	r := &DashboardRun{
		// create RuntimeDependencyPublisherImpl- this handles 'with' run creation and runtime dependency resolution
		RuntimeDependencyPublisherImpl: *NewRuntimeDependencyPublisherImpl(dashboard, executionTree, executionTree),
		parent:                         parent,
		dashboard:                      dashboard,
	}

	// set inputs map on RuntimeDependencyPublisherImpl BEFORE creating child runs
	r.inputs = dashboard.GetInputs()

	// after setting inputs, init runtime dependencies. this creates with runs and adds them to our children
	err := r.initRuntimeDependencies()
	if err != nil {
		return nil, err
	}

	err = r.createChildRuns(executionTree)
	if err != nil {
		return nil, err
	}

	// add r into execution tree
	executionTree.runs[r.Name] = r

	// create buffered channel for children to report their completion
	r.createChildCompleteChan()

	return r, nil
}

// Initialise implements DashboardRunNode
func (r *DashboardRun) Initialise(ctx context.Context) {
	// initialise our children
	if err := r.initialiseChildren(ctx); err != nil {
		r.SetError(ctx, err)
	}
}

// Execute implements DashboardRunNode
// execute all children and wait for them to complete
func (r *DashboardRun) Execute(ctx context.Context) {
	r.executeChildrenAsync(ctx)

	// wait for children to complete
	err := <-r.waitForChildren()
	log.Printf("[TRACE] Execute run %s all children complete, error: %v", r.Name, err)

	if err == nil {
		// set complete status on dashboard
		r.SetComplete(ctx)
	} else {
		r.SetError(ctx, err)
	}
}

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardRun) IsSnapshotPanel() {}

// SetError implements DashboardTreeRun
// tell parent we are done
func (r *DashboardRun) SetError(_ context.Context, err error) {
	r.err = err
	// error type does not serialise to JSON so copy into a string
	r.ErrorString = err.Error()
	r.Status = dashboardtypes.DashboardRunError
	// raise container error event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.DashboardError{
		Dashboard:   r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
	})
	r.parent.ChildCompleteChan() <- r
}

// SetComplete implements DashboardTreeRun
func (r *DashboardRun) SetComplete(context.Context) {
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

// GetInput searches for an input with the given name
func (r *DashboardRun) GetInput(name string) (*modconfig.DashboardInput, bool) {
	return r.dashboard.GetInput(name)
}

// GetInputsDependingOn returns a list o DashboardInputs which have a runtime dependency on the given input
func (r *DashboardRun) GetInputsDependingOn(changedInputName string) []string {
	var res []string
	for _, input := range r.dashboard.Inputs {
		if input.DependsOnInput(changedInputName) {
			res = append(res, input.UnqualifiedName)
		}
	}
	return res
}

func (r *DashboardRun) createChildRuns(executionTree *DashboardExecutionTree) error {
	// ask our resource for its children
	children := r.dashboard.GetChildren()

	for _, child := range children {
		var childRun dashboardtypes.DashboardTreeRun
		var err error
		switch i := child.(type) {
		case *modconfig.DashboardWith:
			// ignore as with runs are created by RuntimeDependencyPublisherImpl
			continue
		case *modconfig.Dashboard:
			childRun, err = NewDashboardRun(i, r, executionTree)
			if err != nil {
				return err
			}
		case *modconfig.DashboardContainer:
			childRun, err = NewDashboardContainerRun(i, r, executionTree)
			if err != nil {
				return err
			}
		case *modconfig.Benchmark, *modconfig.Control:
			childRun, err = NewCheckRun(i.(modconfig.DashboardLeafNode), r, executionTree)
			if err != nil {
				return err
			}
		case *modconfig.DashboardInput:
			// NOTE: clone the input to avoid mutating the original
			// TODO remove the need for this when we refactor input values resolution
			// TODO https://github.com/turbot/steampipe/issues/2864
			childRun, err = NewLeafRun(i.Clone(), r, executionTree)
			if err != nil {
				return err
			}

		default:
			// ensure this item is a DashboardLeafNode
			leafNode, ok := i.(modconfig.DashboardLeafNode)
			if !ok {
				return fmt.Errorf("child %s does not implement DashboardLeafNode", i.Name())
			}

			childRun, err = NewLeafRun(leafNode, r, executionTree)
			if err != nil {
				return err
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
	return nil
}
