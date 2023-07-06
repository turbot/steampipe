package dashboardexecute

import (
	"context"
	"fmt"
	"log"

	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// DashboardRun is a struct representing a container run
type DashboardRun struct {
	runtimeDependencyPublisherImpl

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

func NewDashboardRun(dashboard *modconfig.Dashboard, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) (*DashboardRun, error) {
	r := &DashboardRun{
		parent:    parent,
		dashboard: dashboard,
	}
	// create RuntimeDependencyPublisherImpl- this handles 'with' run creation and resolving runtime dependency resolution
	// (we must create after creating the run as it requires a ref to the run)
	r.runtimeDependencyPublisherImpl = newRuntimeDependencyPublisherImpl(dashboard, parent, r, executionTree)
	// add r into execution tree BEFORE creating child runs or initialising runtime depdencies
	// - this is so child runs can find this dashboard run
	executionTree.runs[r.Name] = r

	// set inputs map on RuntimeDependencyPublisherImpl BEFORE creating child runs
	r.inputs = dashboard.GetInputs()

	// after setting inputs, init runtime dependencies. this creates with runs and adds them to our children
	err := r.initWiths()
	if err != nil {
		return nil, err
	}

	err = r.createChildRuns(executionTree)
	if err != nil {
		return nil, err
	}

	// create buffered channel for children to report their completion
	r.createChildCompleteChan()

	return r, nil
}

// Initialise implements DashboardTreeRun
func (r *DashboardRun) Initialise(ctx context.Context) {
	// initialise our children
	if err := r.initialiseChildren(ctx); err != nil {
		r.SetError(ctx, err)
	}
}

// Execute implements DashboardTreeRun
// execute all children and wait for them to complete
func (r *DashboardRun) Execute(ctx context.Context) {
	r.executeChildrenAsync(ctx)

	// try to set status as running (will be set to blocked if any children are blocked)
	r.setRunning(ctx)

	// wait for children to complete
	err := <-r.waitForChildrenAsync(ctx)
	if err == nil {
		log.Printf("[TRACE] Execute run %s all children complete, success", r.Name)
		// set complete status on dashboard
		r.SetComplete(ctx)
	} else {
		log.Printf("[TRACE] Execute run %s all children complete, error: %s", r.Name, err.Error())
		r.SetError(ctx, err)
	}
}

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardRun) IsSnapshotPanel() {}

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

			// TACTICAL: as this is a runtime dependency,  set the run name to the 'scoped name'
			// this is to match the name in the panel dependendencies
			// TODO [node_reuse] consider naming https://github.com/turbot/steampipe/issues/2921
			inputRunName := fmt.Sprintf("%s.%s", r.DashboardName, i.UnqualifiedName)
			childRun, err = NewLeafRun(i.Clone(), r, executionTree, setName(inputRunName))
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
		if childRun.GetRunStatus() == dashboardtypes.RunInitialized {
			r.Status = dashboardtypes.RunInitialized
		}
		r.children = append(r.children, childRun)
	}
	return nil
}
