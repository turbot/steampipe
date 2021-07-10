package controlexecute

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"

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
	enabled  bool
}

func NewControlProgressRenderer(total int) *ControlProgressRenderer {
	return &ControlProgressRenderer{
		total:   total,
		pending: total,
		enabled: viper.GetBool(constants.ArgProgress)}
}

func (p *ControlProgressRenderer) Start() {
	if p.enabled {
		p.spinner = display.ShowSpinner("")
	}
}
func (p *ControlProgressRenderer) OnControlStart(control *modconfig.Control) {
	if p.enabled {
		p.current = typehelpers.SafeString(control.Title)
		display.UpdateSpinnerMessage(p.spinner, p.message())
	}
}
func (p *ControlProgressRenderer) OnComplete() {
	if p.enabled {
		p.pending--
		p.complete++
		display.UpdateSpinnerMessage(p.spinner, p.message())
	}
}

func (p *ControlProgressRenderer) OnError() {
	if p.enabled {
		p.pending--
		p.error++
		display.UpdateSpinnerMessage(p.spinner, p.message())
	}
}

func (p *ControlProgressRenderer) Finish() {
	if p.enabled {
		display.StopSpinner(p.spinner)
	}
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
