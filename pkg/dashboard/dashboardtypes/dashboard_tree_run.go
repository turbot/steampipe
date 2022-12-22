package dashboardtypes

import (
	"context"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// DashboardTreeRun is an interface implemented by all dashboard run nodes
type DashboardTreeRun interface {
	Initialise(ctx context.Context)
	Execute(ctx context.Context)
	GetName() string
	GetTitle() string
	GetRunStatus() DashboardRunStatus
	SetError(context.Context, error)
	GetError() error
	GetParent() DashboardParent
	SetComplete(context.Context)
	RunComplete() bool
	GetInputsDependingOn(string) []string
	GetNodeType() string
	AsTreeNode() *SnapshotTreeNode
	GetResource() modconfig.DashboardLeafNode
}

type TreeRunExecuteConfig struct {
	RuntimeDependenciesOnly bool
	BaseExecution           bool
}
