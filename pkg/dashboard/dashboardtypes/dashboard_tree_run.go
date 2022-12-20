package dashboardtypes

import (
	"context"
)

// DashboardTreeRun is an interface implemented by all dashboard run nodes
type DashboardTreeRun interface {
	Initialise(ctx context.Context)
	Execute(ctx context.Context, opts ...TreeRunExecuteOption)
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
}

type TreeRunExecuteOption = func(target *TreeRunExecuteConfig)
type TreeRunExecuteConfig struct {
	RuntimeDepedenciesOnly bool
	BaseExecution          bool
}

func RuntimeDependenciesOnly() TreeRunExecuteOption {
	return func(target *TreeRunExecuteConfig) {
		target.RuntimeDepedenciesOnly = true
	}
}
func BaseExecution() TreeRunExecuteOption {
	return func(target *TreeRunExecuteConfig) {
		target.BaseExecution = true
		target.RuntimeDepedenciesOnly = true
	}
}
