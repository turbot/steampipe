package reportinterfaces

import (
	"context"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

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
	GetPath() modconfig.NodePath
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
