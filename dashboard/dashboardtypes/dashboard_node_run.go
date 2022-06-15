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
	GetRunStatus() DashboardRunStatus
	SetError(err error)
	GetError() error
	SetComplete()
	RunComplete() bool
	GetChildren() []DashboardNodeRun
	ChildrenComplete() bool
	GetInputsDependingOn(changedInputName string) []string
	AsTreeNode() *SnapshotTreeNode
}

// DashboardNodeParent is an interface implemented by all dashboard run nodes which have children
type DashboardNodeParent interface {
	GetName() string
	ChildCompleteChan() chan DashboardNodeRun
}
