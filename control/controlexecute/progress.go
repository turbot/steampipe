package controlexecute

import (
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

	// status update hooks
	statusHook statushooks.StatusHooks
}

func NewControlProgressRenderer(total int) *ControlProgressRenderer {
	return &ControlProgressRenderer{
		updateLock: &sync.Mutex{},
		total:      total,
		pending:    total,
		enabled:    viper.GetBool(constants.ArgProgress),
	}
}

func (p *ControlProgressRenderer) Start() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	if p.enabled {
		p.statusHook.SetStatus("Starting controls...")
	}
}

func (p *ControlProgressRenderer) OnControlStart(control *modconfig.Control) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	// increment the parallel execution count
	p.executing++

	// decrement pending count
	p.pending--

	if p.enabled {
		p.statusHook.SetStatus(p.message())
	}
}

func (p *ControlProgressRenderer) OnControlFinish() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	// decrement the parallel execution count
	p.executing--
	if p.enabled {
		p.statusHook.SetStatus(p.message())
	}
}

func (p *ControlProgressRenderer) OnControlComplete() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	p.complete++

	if p.enabled {
		p.statusHook.SetStatus(p.message())
	}
}

func (p *ControlProgressRenderer) OnControlError() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	p.error++

	if p.enabled {
		p.statusHook.SetStatus(p.message())
	}
}

func (p *ControlProgressRenderer) Finish() {
	if p.enabled {
		p.statusHook.Done()
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
