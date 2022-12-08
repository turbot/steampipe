package dashboardtypes

type DashboardRunStatus string

const (
	DashboardRunReady    DashboardRunStatus = "ready"
	DashboardRunComplete DashboardRunStatus = "complete"
	DashboardRunError    DashboardRunStatus = "error"
)
