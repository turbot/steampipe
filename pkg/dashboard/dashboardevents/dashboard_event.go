package dashboardevents

import "context"

type DashboardEvent interface {
	IsDashboardEvent()
}
type DashboardEventHandler func(context.Context, DashboardEvent)
