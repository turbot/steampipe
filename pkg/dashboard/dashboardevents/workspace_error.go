package dashboardevents

type WorkspaceError struct {
	Error error
}

// IsDashboardEvent implements DashboardEvent interface
func (*WorkspaceError) IsDashboardEvent() {}
