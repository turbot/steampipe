package controlexecute

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/briandowns/spinner"
	"github.com/turbot/steampipe/display"
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
		enabled:    viper.GetBool(constants.ArgProgress)}
}

func (p *ControlProgressRenderer) Start() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	if p.enabled {
		p.spinner = display.ShowSpinner("")
	}
}

func (p *ControlProgressRenderer) OnControlExecuteStart() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	if p.enabled {
		// increment the parallel execution count
		p.executing++
		display.UpdateSpinnerMessage(p.spinner, p.message())
	}
}

func (p *ControlProgressRenderer) OnControlExecuteFinish() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()
	if p.enabled {
		// decrement the parallel execution count
		p.executing--
		display.UpdateSpinnerMessage(p.spinner, p.message())
	}
}

func (p *ControlProgressRenderer) OnControlStart(control *modconfig.Control) {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	if p.enabled {
		display.UpdateSpinnerMessage(p.spinner, p.message())
	}
}
func (p *ControlProgressRenderer) OnControlComplete() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	if p.enabled {
		p.pending--
		p.complete++
		display.UpdateSpinnerMessage(p.spinner, p.message())
	}
}

func (p *ControlProgressRenderer) OnControlError() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	if p.enabled {
		p.pending--
		p.error++
		display.UpdateSpinnerMessage(p.spinner, p.message())
	}
}

func (p *ControlProgressRenderer) Finish() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	if p.enabled {
		display.StopSpinner(p.spinner)
	}
}

func (p ControlProgressRenderer) message() string {
	return fmt.Sprintf("Running %d %s. (%d complete, %d pending, %d %s) [%d in parallel]",
		p.total,
		utils.Pluralize("control", p.total),
		p.complete,
		p.pending,
		p.error,
		utils.Pluralize("error", p.error),
		p.executing,
	)
}
