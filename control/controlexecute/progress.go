package controlexecute

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/briandowns/spinner"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/utils"
)

type ControlProgressRenderer struct {
	startTime          time.Time
	timeToFirstControl time.Duration
	updateLock         *sync.Mutex
	total              int
	pending            int
	complete           int
	error              int
	spinner            *spinner.Spinner
	enabled            bool
	executing          int
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

	p.startTime = time.Now()

	if p.enabled {
		p.spinner = display.ShowSpinner("Starting Controls")
	}
}

func (p *ControlProgressRenderer) OnControlExecuteStart() {
	p.updateLock.Lock()
	defer p.updateLock.Unlock()

	if p.timeToFirstControl == 0 {
		p.timeToFirstControl = time.Since(p.startTime)
		log.Println("[WARN] Time to first control:", p.timeToFirstControl)
	}

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
