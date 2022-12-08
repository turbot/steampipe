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

// DashboardRun is a struct representing a container run
type DashboardRun struct {
	RuntimeDependencyPublisherBase
	DashboardParentBase
	Width            int               `json:"width,omitempty"`
	Description      string            `json:"description,omitempty"`
	Display          string            `json:"display,omitempty"`
	Documentation    string            `json:"documentation,omitempty"`
	Tags             map[string]string `json:"tags,omitempty"`
	ErrorString      string            `json:"error,omitempty"`
	NodeType         string            `json:"panel_type"`
	DashboardName    string            `json:"dashboard"`
	SourceDefinition string            `json:"source_definition"`

	dashboardNode *modconfig.Dashboard
	parent        dashboardtypes.DashboardParent
}

func (r *DashboardRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
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

func NewDashboardRun(dashboard *modconfig.Dashboard, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) (*DashboardRun, error) {
	r := &DashboardRun{
		RuntimeDependencyPublisherBase: NewRuntimeDependencyPublisherBase(dashboard, parent, executionTree),
		NodeType:                       modconfig.BlockTypeDashboard,
		DashboardName:                  executionTree.dashboardName,
		Description:                    typehelpers.SafeString(dashboard.Description),
		Display:                        typehelpers.SafeString(dashboard.Display),
		Documentation:                  typehelpers.SafeString(dashboard.Documentation),
		Tags:                           dashboard.Tags,
		SourceDefinition:               dashboard.GetMetadata().SourceDefinition,
		parent:                         parent,
		dashboardNode:                  dashboard,
	}
	if dashboard.Width != nil {
		r.Width = *dashboard.Width
	}

	// if we have with blocks, create runs for them
	// BEFORE creating child runs, and before adding runtime dependencies
	withBlocks := dashboard.GetWiths()
	if len(withBlocks) > 0 {
		// create the child runs
		err := r.createWithRuns(withBlocks, executionTree)
		if err != nil {
			return nil, err
		}
	}

	if err := r.resolveRuntimeDependencies(dashboard); err != nil {
		return nil, err
	}

	// set inputs map on RuntimeDependencyPublisherBase BEFORE creating child runs
	r.inputs = dashboard.GetInputs()

	children := dashboard.GetChildren()
	err := r.createChildRuns(children, executionTree)
	if err != nil {
		return nil, err
	}

	// add r into execution tree
	executionTree.runs[r.Name] = r

	// create buffered child complete chan
	if childCount := len(children) + len(withBlocks); childCount > 0 {
		r.childComplete = make(chan dashboardtypes.DashboardTreeRun, childCount)
	}

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
	// TODO use with same execution as leaf runs
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
		// fall through to recheck ChildrenCompletes
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
	return r.dashboardNode.GetInput(name)
}

// GetInputsDependingOn returns a list o DashboardInputs which have a runtime dependency on the given input
func (r *DashboardRun) GetInputsDependingOn(changedInputName string) []string {
	var res []string
	for _, input := range r.dashboardNode.Inputs {
		if input.DependsOnInput(changedInputName) {
			res = append(res, input.UnqualifiedName)
		}
	}
	return res
}

func (r *DashboardRun) createWithRuns(withs []*modconfig.DashboardWith, executionTree *DashboardExecutionTree) error {
	for _, w := range withs {
		withRun, err := NewLeafRun(w, r, executionTree)
		if err != nil {
			return err
		}
		r.withRuns[w.UnqualifiedName] = withRun
	}
	return nil
}

func (r *DashboardRun) createChildRuns(children []modconfig.ModTreeItem, executionTree *DashboardExecutionTree) error {
	for _, child := range children {
		var childRun dashboardtypes.DashboardTreeRun
		var err error
		switch i := child.(type) {
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
			// TODO [reports] remove the need for this when we refactor input values resolution
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
