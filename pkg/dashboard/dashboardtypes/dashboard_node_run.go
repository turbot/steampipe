package dashboardtypes

import (
	"context"
)

type DashboardRunStatus string

const (
	DashboardRunReady    DashboardRunStatus = "ready"
	DashboardRunComplete DashboardRunStatus = "complete"
	DashboardRunError    DashboardRunStatus = "error"
)

// DashboardNodeRun is an interface implemented by all dashboard run nodes
type DashboardNodeRun interface {
	Initialise(ctx context.Context)
	Execute(ctx context.Context)
	GetName() string
	GetTitle() string
	GetRunStatus() DashboardRunStatus
	SetError(context.Context, error)
	GetError() error
	SetComplete(context.Context)
	RunComplete() bool
	GetChildren() []DashboardNodeRun
	ChildrenComplete() bool
	GetInputsDependingOn(string) []string
	AsTreeNode() *SnapshotTreeNode
}

// DashboardNodeParent is an interface implemented by all dashboard run nodes which have children
type DashboardNodeParent interface {
	GetName() string
	ChildCompleteChan() chan DashboardNodeRun
}
