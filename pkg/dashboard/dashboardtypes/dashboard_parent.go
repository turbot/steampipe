package dashboardtypes

import "context"

// DashboardParent is an interface implemented by all dashboard run nodes which have children
type DashboardParent interface {
	DashboardTreeRun
	GetName() string
	ChildCompleteChan() chan DashboardTreeRun
	GetChildren() []DashboardTreeRun
	ChildrenComplete() bool
	ChildStatusChanged(context.Context)
}
