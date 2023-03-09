package dashboardevents

import "time"

// ExecutionError is an event which is sent if an error occusrs _before execution has started_
// e.g. a failure to create the execution tree
type ExecutionError struct {
	Error     error
	Session   string
	Timestamp time.Time
}

// IsDashboardEvent implements DashboardEvent interface
func (*ExecutionError) IsDashboardEvent() {}
