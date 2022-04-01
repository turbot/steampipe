package controlstatus

import (
	"context"
	"sync"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ControlProgress struct {
	updateLock      *sync.Mutex
	Total           int            `json:"total"`
	Pending         int            `json:"pending"`
	Complete        int            `json:"complete"`
	Error           int            `json:"error"`
	Executing       int            `json:"executing"`
	StatusSummaries *StatusSummary `json:"control_row_status_summary"`
}

func NewControlProgress(total int) *ControlProgress {
	return &ControlProgress{
		updateLock:      &sync.Mutex{},
		Total:           total,
		Pending:         total,
		StatusSummaries: &StatusSummary{},
	}
}

func (p *ControlProgress) Start(ctx context.Context) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	OnStart(ctx, p)
}

func (p *ControlProgress) OnControlStart(ctx context.Context, control *modconfig.Control) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	// increment the parallel execution count
	p.Executing++

	// decrement pending count
	p.Pending--

	OnControlStart(ctx, control, p)
}

func (p *ControlProgress) OnControlComplete(ctx context.Context, control *modconfig.Control, controlRunStatus ControlRunStatus, controlStatusSummary *StatusSummary) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	p.Complete++
	// decrement the parallel execution count
	p.Executing--
	p.StatusSummaries.Merge(controlStatusSummary)
	OnControlComplete(ctx, control, controlRunStatus, controlStatusSummary, p)
}

func (p *ControlProgress) OnControlError(ctx context.Context, control *modconfig.Control, controlRunStatus ControlRunStatus, controlStatusSummary *StatusSummary) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	p.Error++
	// decrement the parallel execution count
	p.Executing--
	p.StatusSummaries.Merge(controlStatusSummary)
	OnControlError(ctx, control, controlRunStatus, controlStatusSummary, p)
}

func (p *ControlProgress) Finish(ctx context.Context) {
	OnComplete(ctx, p)
}
