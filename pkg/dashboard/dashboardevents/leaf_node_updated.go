package dashboardevents

import (
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"time"
)

type LeafNodeUpdated struct {
	LeafNode    map[string]any
	Session     string
	ExecutionId string
	Timestamp   time.Time
}

func NewLeafNodeUpdate(r dashboardtypes.DashboardTreeRun, session, executionId string) (*LeafNodeUpdated, error) {
	immutableNode, err := utils.JsonCloneToMap(r)
	if err != nil {
		return nil, err
	}
	return &LeafNodeUpdated{
		LeafNode:    immutableNode,
		Session:     session,
		ExecutionId: executionId,
		Timestamp:   time.Now(),
	}, nil
}

// IsDashboardEvent implements DashboardEvent interface
func (*LeafNodeUpdated) IsDashboardEvent() {}
