package dashboardtypes

type RunStatus string

const (
	RunInitialized RunStatus = "initialized"
	RunBlocked     RunStatus = "blocked"
	RunRunning     RunStatus = "running"
	RunComplete    RunStatus = "complete"
	RunError       RunStatus = "error"
	RunCanceled    RunStatus = "canceled"
)

func (s RunStatus) IsError() bool {
	return s == RunError || s == RunCanceled
}

func (s RunStatus) IsFinished() bool {
	return s == RunComplete || s.IsError()
}
