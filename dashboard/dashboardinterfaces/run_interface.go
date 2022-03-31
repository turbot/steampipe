package dashboardinterfaces

import (
	"context"
)

type DashboardRunStatus string

// TODO [report] think about status - do we need in progress
const (
	DashboardRunReady    DashboardRunStatus = "ready"
	DashboardRunBlocked                     = "blocked"
	DashboardRunComplete                    = "complete"
	DashboardRunError                       = "error"
)

type DashboardNodeRun interface {
	Execute(ctx context.Context)
	GetName() string
	GetRunStatus() DashboardRunStatus
	SetError(err error)
	GetError() error
	SetComplete()
	RunComplete() bool
	ChildrenComplete() bool
	GetInputsDependingOn(changedInputName string) []string
}

type DashboardNodeParent interface {
	GetName() string
	ChildCompleteChan() chan DashboardNodeRun
}
