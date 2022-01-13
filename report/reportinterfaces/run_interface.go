package reportinterfaces

import "context"

type ReportRunStatus uint32

const (
	ReportRunReady ReportRunStatus = 1 << iota
	ReportRunStarted
	ReportRunComplete
	ReportRunError
)

type ReportNodeRun interface {
	Execute(ctx context.Context) error
	GetName() string
	GetRunStatus() ReportRunStatus
	SetError(err error)
	SetComplete()
	RunComplete() bool
	ChildrenComplete() bool
}

type ReportNodeParent interface {
	GetName() string
	ChildCompleteChan() chan ReportNodeRun
}
