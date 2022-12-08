package dashboardexecute

import (
	"context"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
)

type DashboardParentBase struct {
	children      []dashboardtypes.DashboardTreeRun
	childComplete chan dashboardtypes.DashboardTreeRun
}

func (r *DashboardParentBase) initialiseChildren(ctx context.Context) error {
	var errors []error
	for _, child := range r.children {
		child.Initialise(ctx)
		if err := child.GetError(); err != nil {
			errors = append(errors, err)
		}
	}

	return error_helpers.CombineErrors(errors...)

}

// GetChildren implements DashboardTreeRun
func (r *DashboardParentBase) GetChildren() []dashboardtypes.DashboardTreeRun {
	return r.children
}

// ChildrenComplete implements DashboardTreeRun
func (r *DashboardParentBase) ChildrenComplete() bool {
	for _, child := range r.children {
		if !child.RunComplete() {
			return false
		}
	}

	return true
}

func (r *DashboardParentBase) ChildCompleteChan() chan dashboardtypes.DashboardTreeRun {
	return r.childComplete
}
