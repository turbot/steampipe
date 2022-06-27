package controlstatus

// ControlRunStatusProvider is an interface used to allow us to pass a control as the payload of ControlComplete and ControlError events -
type ControlRunStatusProvider interface {
	GetControlId() string
	GetRunStatus() ControlRunStatus
	GetStatusSummary() *StatusSummary
}
