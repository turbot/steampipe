package dashboardexecute

import (
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type DashboardTreeRunBase struct {
	Name   string                            `json:"name"`
	Title  string                            `json:"title,omitempty"`
	Status dashboardtypes.DashboardRunStatus `json:"status"`

	err           error
	parent        dashboardtypes.DashboardParent
	executionTree *DashboardExecutionTree
}

func NewDashboardTreeRunBase(resource modconfig.HclResource, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) DashboardTreeRunBase {
	// NOTE: for now we MUST declare children inline - therefore we cannot share children between runs in the tree
	// (if we supported the children property then we could reuse resources)
	// so FOR NOW it is safe to use the container name directly as the run name
	return DashboardTreeRunBase{
		Name:  resource.Name(),
		Title: resource.GetTitle(),
		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		Status:        dashboardtypes.DashboardRunComplete,
		parent:        parent,
		executionTree: executionTree,
	}

}

// GetName implements DashboardTreeRun
func (r *DashboardTreeRunBase) GetName() string {
	return r.Name
}

// GetTitle implements DashboardTreeRun
func (r *DashboardTreeRunBase) GetTitle() string {
	return r.Title
}

// GetRunStatus implements DashboardTreeRun
func (r *DashboardTreeRunBase) GetRunStatus() dashboardtypes.DashboardRunStatus {
	return r.Status
}

// GetError implements DashboardTreeRun
func (r *DashboardTreeRunBase) GetError() error {
	return r.err
}

// RunComplete implements DashboardTreeRun
func (r *DashboardTreeRunBase) RunComplete() bool {
	return r.Status == dashboardtypes.DashboardRunComplete || r.Status == dashboardtypes.DashboardRunError
}

// GetInputsDependingOn implements DashboardTreeRun
// defaults to nothing
func (r *DashboardTreeRunBase) GetInputsDependingOn(_ string) []string { return nil }

// GetParent implements DashboardTreeRun
func (r *DashboardTreeRunBase) GetParent() dashboardtypes.DashboardParent {
	return r.parent
}
