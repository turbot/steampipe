package dashboardevents

type ExecutionError struct {
	Error   error
	Session string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionError) IsDashboardEvent() {}
