package dashboardtypes

type DashboardRunStatus string

const (
	DashboardRunReady    DashboardRunStatus = "ready"
	DashboardRunBlocked  DashboardRunStatus = "blocked"
	DashboardRunRunning  DashboardRunStatus = "running"
	DashboardRunComplete DashboardRunStatus = "complete"
	DashboardRunError    DashboardRunStatus = "error"
)
