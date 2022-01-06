package controlexecute

import (
	"context"
	"fmt"
	"sync"

	"github.com/turbot/steampipe/statushooks"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/briandowns/spinner"
	"github.com/turbot/steampipe/utils"
)

type ControlProgressRenderer struct {
	updateLock *sync.Mutex
	total      int
	pending    int
	complete   int
	error      int
	spinner    *spinner.Spinner
	enabled    bool
	executing  int
}

func NewControlProgressRenderer(total int) *ControlProgressRenderer {
	return &ControlProgressRenderer{
		updateLock: &sync.Mutex{},
		total:      total,
		pending:    total,
		enabled:    viper.GetBool(constants.ArgProgress),
	}
}

func (p *ControlProgressRenderer) Start(ctx context.Context) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	if p.enabled {
		statushooks.SetStatus(ctx, "Starting controls...")
	}
}

func (p *ControlProgressRenderer) OnControlStart(ctx context.Context, control *modconfig.Control) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	// increment the parallel execution count
	p.executing++

	// decrement pending count
	p.pending--

	if p.enabled {
		statushooks.SetStatus(ctx, p.message())
	}
}

func (p *ControlProgressRenderer) OnControlFinish(ctx context.Context) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	// decrement the parallel execution count
	p.executing--
	if p.enabled {
		statushooks.SetStatus(ctx, p.message())
	}
}

func (p *ControlProgressRenderer) OnControlComplete(ctx context.Context) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	p.complete++

	if p.enabled {
		statushooks.SetStatus(ctx, p.message())
	}
}

func (p *ControlProgressRenderer) OnControlError(ctx context.Context) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	p.error++

	if p.enabled {
		statushooks.SetStatus(ctx, p.message())
	}
}

func (p *ControlProgressRenderer) Finish(ctx context.Context) {
	if p.enabled {
		statushooks.Done(ctx)
	}
}

func (p ControlProgressRenderer) message() string {
	return fmt.Sprintf("Running %d %s. (%d complete, %d running, %d pending, %d %s)",
		p.total,
		utils.Pluralize("control", p.total),
		p.complete,
		p.executing,
		p.pending,
		p.error,
		utils.Pluralize("error", p.error),
	)
}
