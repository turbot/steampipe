package dashboardtypes

type DashboardRunStatus string

const (
	DashboardRunInitialized DashboardRunStatus = "initialized"
	DashboardRunBlocked     DashboardRunStatus = "blocked"
	DashboardRunRunning     DashboardRunStatus = "running"
	DashboardRunComplete    DashboardRunStatus = "complete"
	DashboardRunError       DashboardRunStatus = "error"
	DashboardRunCanceled    DashboardRunStatus = "canceled"
)

func (s DashboardRunStatus) IsError() bool {
	return s == DashboardRunError || s == DashboardRunCanceled
}
