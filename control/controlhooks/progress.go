package controlhooks

import (
	"context"
	"sync"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ControlProgress struct {
	updateLock *sync.Mutex
	Total      int `json:"total"`
	Pending    int `json:"pending"`
	Complete   int `json:"complete"`
	Error      int `json:"error"`
	Executing  int `json:"executing"`
}

func NewControlProgress(total int) *ControlProgress {
	return &ControlProgress{
		updateLock: &sync.Mutex{},
		Total:      total,
		Pending:    total,
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

	OnControlEvent(ctx, p)
}

func (p *ControlProgress) OnControlFinish(ctx context.Context) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	// decrement the parallel execution count
	p.Executing--

	OnControlEvent(ctx, p)
}

func (p *ControlProgress) OnControlComplete(ctx context.Context) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	p.Complete++

	OnControlEvent(ctx, p)
}

func (p *ControlProgress) OnControlError(ctx context.Context) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	p.Error++

	OnControlEvent(ctx, p)
}

func (p *ControlProgress) Finish(ctx context.Context) {
	OnDone(ctx, p)
}
