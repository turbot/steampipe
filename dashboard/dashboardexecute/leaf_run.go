package dashboardexecute

import (
	"context"

	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// LeafRun is a struct representing the execution of a leaf dashboard node
type LeafRun struct {
	Name string `json:"name"`

	Title         string                   `json:"title,omitempty"`
	Width         int                      `json:"width,omitempty"`
	SQL           string                   `json:"sql,omitempty"`
	Data          *LeafData                `json:"data,omitempty"`
	Error         error                       `json:"error,omitempty"`
	DashboardNode modconfig.DashboardLeafNode `json:"properties"`
	NodeType      string   `json:"node_type"`
	DashboardName string   `json:"dashboard"`
	Path          []string `json:"-"`
	parent        dashboardinterfaces.DashboardNodeParent
	runStatus     dashboardinterfaces.DashboardRunStatus
	executionTree *DashboardExecutionTree
}

func NewLeafRun(resource modconfig.DashboardLeafNode, parent dashboardinterfaces.DashboardNodeParent, executionTree *DashboardExecutionTree) (*LeafRun, error) {
	// ensure the tree node name is unique
	name := executionTree.GetUniqueName(resource.Name())

	r := &LeafRun{
		Name:          name,
		Title:         resource.GetTitle(),
		Width:         resource.GetWidth(),
		SQL:           resource.GetSQL(),
		Path:          resource.GetPaths()[0],
		DashboardNode: resource,
		DashboardName: executionTree.dashboardName,
		executionTree: executionTree,
		parent:        parent,

		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		runStatus: dashboardinterfaces.DashboardRunComplete,
	}

	parsedName, err := modconfig.ParseResourceName(resource.Name())
	if err != nil {
		return nil, err
	}
	r.NodeType = parsedName.ItemType
	// if we have sql, set status to ready
	if r.SQL != "" {
		r.runStatus = dashboardinterfaces.DashboardRunReady
	}

	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r, nil
}

// Execute implements DashboardRunNode
func (r *LeafRun) Execute(ctx context.Context) error {
	// if there are any unresolved runtime dependencies, wait for them
	if err := r.waitForRuntimeDependencies(ctx); err != nil {
		return err
	}

	if r.SQL == "" {
		return nil
	}

	var err error
	queryResult, err := r.executionTree.client.ExecuteSync(ctx, r.SQL)
	if err != nil {
		// set the error status on the counter - this will raise counter error event
		r.SetError(err)
		return err

	}
	r.Data = NewLeafData(queryResult)
	// set complete status on counter - this will raise counter complete event
	r.SetComplete()
	return nil
}

// GetName implements DashboardNodeRun
func (r *LeafRun) GetName() string {
	return r.Name
}

// GetPath implements DashboardNodeRun
func (r *LeafRun) GetPath() modconfig.NodePath {
	return r.Path
}

// GetRunStatus implements DashboardNodeRun
func (r *LeafRun) GetRunStatus() dashboardinterfaces.DashboardRunStatus {
	return r.runStatus
}

// SetError implements DashboardNodeRun
func (r *LeafRun) SetError(err error) {
	r.Error = err
	r.runStatus = dashboardinterfaces.DashboardRunError
	// raise counter error event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeError{Node: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r

}

// SetComplete implements DashboardNodeRun
func (r *LeafRun) SetComplete() {
	r.runStatus = dashboardinterfaces.DashboardRunComplete
	// raise counter complete event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeComplete{Node: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements DashboardNodeRun
func (r *LeafRun) RunComplete() bool {
	return r.runStatus == dashboardinterfaces.DashboardRunComplete || r.runStatus == dashboardinterfaces.DashboardRunError
}

// ChildrenComplete implements DashboardNodeRun
func (r *LeafRun) ChildrenComplete() bool {
	return true
}

func (r *LeafRun) waitForRuntimeDependencies(ctx context.Context) error {
	runtimeDependencies := r.DashboardNode.GetRuntimeDependencies()

	// runtime dependencies are always (for now) dashboard inputs

	for _, dependency := range runtimeDependencies {
		// check with the top level dashboard whether the dependency is available
		inputValue, err := r.executionTree.Root.GetRuntimeDependency(dependency)
		if err != nil {
			return err
		}
		if inputValue != nil {
			// ok we have this one - set it on the dependency
			dependency.Value = inputValue
			continue
		}

		if err := r.executionTree.waitForRuntimeDependency(ctx, dependency); err != nil {
			return err
		}

		// now populate the runtime dependency target property
		//r.setRuntimeDependency()
	}

	return nil
}
