package dashboardevents

// ExecutionError is an event which is sent if an error occusrs _before execution has started_
// e.g. a failure to create the execution tree
type ExecutionError struct {
	Error   error
	Session string
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionError) IsDashboardEvent() {}
