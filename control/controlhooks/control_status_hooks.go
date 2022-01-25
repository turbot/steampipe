package controlhooks

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/utils"
)

// ControlStatusHooks is a struct which implements ControlHooks, and displays the control progress as a status message
type ControlStatusHooks struct {
	Enabled bool
}

func NewControlStatusHooks() *ControlStatusHooks {
	return &ControlStatusHooks{
		Enabled: viper.GetBool(constants.ArgProgress),
	}
}

func (c *ControlStatusHooks) OnControlEvent(ctx context.Context, p *ControlProgress) {
	if !c.Enabled {
		return
	}

	var message string
	if p.Total == 0 {
		message = "Starting controls..."
	} else {
		message = fmt.Sprintf("Running %d %s. (%d complete, %d running, %d pending, %d %s)",
			p.Total,
			utils.Pluralize("control", p.Total),
			p.Complete,
			p.Executing,
			p.Pending,
			p.Error,
			utils.Pluralize("error", p.Error),
		)
	}
	statushooks.SetStatus(ctx, message)

}
func (c *ControlStatusHooks) OnDone(ctx context.Context, _ *ControlProgress) {
	if !c.Enabled {
		return
	}

	statushooks.Done(ctx)
}
