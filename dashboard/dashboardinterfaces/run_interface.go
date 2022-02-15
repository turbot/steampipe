package dashboardinterfaces

import (
	"context"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type DashboardRunStatus string

const (
	DashboardRunReady    DashboardRunStatus = "ready"
	DashboardRunComplete = "complete"
	DashboardRunError    = "error"
)

type DashboardNodeRun interface {
	Execute(ctx context.Context) error
	GetName() string
	GetPath() modconfig.NodePath
	GetRunStatus() DashboardRunStatus
	SetError(err error)
	SetComplete()
	RunComplete() bool
	ChildrenComplete() bool
}

type DashboardNodeParent interface {
	GetName() string
	ChildCompleteChan() chan DashboardNodeRun
}
