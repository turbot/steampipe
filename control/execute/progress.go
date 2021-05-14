package execute

import (
	"fmt"

	"github.com/briandowns/spinner"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/utils"
)

type ControlProgressRenderer struct {
	total    int
	pending  int
	complete int
	error    int
	spinner  *spinner.Spinner
	current  string
}

func NewControlProgressRenderer(total int) *ControlProgressRenderer {
	return &ControlProgressRenderer{total: total, pending: total}
}

func (p *ControlProgressRenderer) Start() {
	p.spinner = display.ShowSpinner("")
}
func (p *ControlProgressRenderer) OnControlStart(name string) {
	p.current = name
	display.UpdateSpinnerMessage(p.spinner, p.message())
}
func (p *ControlProgressRenderer) OnComplete() {
	p.pending--
	p.complete++
	display.UpdateSpinnerMessage(p.spinner, p.message())
}

func (p *ControlProgressRenderer) OnError() {
	p.pending--
	p.error++
	display.UpdateSpinnerMessage(p.spinner, p.message())
}

func (p *ControlProgressRenderer) Finish() {
	display.StopSpinner(p.spinner)
}

func (p ControlProgressRenderer) message() string {
	return fmt.Sprintf("Running %d %s. (%d complete, %d pending, %d errors): executing \"%s\"",
		p.total,
		utils.Pluralize("control", p.total),
		p.complete,
		p.pending,
		p.error,
		p.current)
}
