package dashboardevents

type DashboardEvent interface {
	IsDashboardEvent()
}
type DashboardEventHandler func(DashboardEvent)
