package dashboardevents

type ExecutionError struct {
	Error       error
	Session     string
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionError) IsDashboardEvent() {}
