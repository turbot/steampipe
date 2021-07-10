package reportinterfaces

type ReportRunStatus uint32

const (
	ReportRunReady ReportRunStatus = 1 << iota
	ReportRunStarted
	ReportRunComplete
	ReportRunError
)

type ReportNodeRun interface {
	GetName() string
	GetRunStatus() ReportRunStatus
	SetError(err error)
	SetComplete()
	RunComplete() bool
	ChildrenComplete() bool
}
