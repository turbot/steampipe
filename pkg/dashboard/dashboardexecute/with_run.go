package dashboardexecute

import (
	"context"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type WithRun struct {
	LeafRun
}

func NewWithRun(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardNodeParent, executionTree *DashboardExecutionTree) (*WithRun, error) {
	r, err := NewLeafRun(resource, parent, executionTree)
	if err != nil {
		return nil, err
	}
	return &WithRun{LeafRun: *r}, nil
}

func (r *WithRun) Execute(ctx context.Context) {
	r.LeafRun.Execute(ctx)
	if r.Status == dashboardtypes.DashboardRunError {
		return
	}
	r.parent.(*LeafRun).setWithValue(r.DashboardNode.GetUnqualifiedName(), r.Data)
}
