package controlstatus

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
)

// SnapshotControlHooks is a struct which implements ControlHooks, and displays the control progress as a status message
type SnapshotControlHooks struct {
	Enabled bool
}

func NewSnapshotControlHooks() *SnapshotControlHooks {
	return &SnapshotControlHooks{
		Enabled: viper.GetBool(constants.ArgProgress),
	}
}

func (c *SnapshotControlHooks) OnStart(ctx context.Context, _ *ControlProgress) {
	if !c.Enabled {
		return
	}

	statushooks.SetStatus(ctx, "Starting...")
}

func (c *SnapshotControlHooks) OnControlStart(ctx context.Context, _ ControlRunStatusProvider, p *ControlProgress) {
	if !c.Enabled {
		return
	}

	c.setStatusFromProgress(ctx, p)
}

func (c *SnapshotControlHooks) OnControlComplete(ctx context.Context, _ ControlRunStatusProvider, p *ControlProgress) {
	if !c.Enabled {
		return
	}

	c.setStatusFromProgress(ctx, p)
}

func (c *SnapshotControlHooks) OnControlError(ctx context.Context, _ ControlRunStatusProvider, p *ControlProgress) {
	if !c.Enabled {
		return
	}

	c.setStatusFromProgress(ctx, p)
}

func (c *SnapshotControlHooks) OnComplete(ctx context.Context, _ *ControlProgress) {
	if !c.Enabled {
		return
	}

	statushooks.Done(ctx)
}

func (c *SnapshotControlHooks) setStatusFromProgress(ctx context.Context, p *ControlProgress) {
	message := fmt.Sprintf("Running %d %s. (%d complete, %d running, %d pending, %d %s)",
		p.Total,
		utils.Pluralize("control", p.Total),
		p.Complete,
		p.Executing,
		p.Pending,
		p.Error,
		utils.Pluralize("error", p.Error),
	)

	statushooks.SetStatus(ctx, message)
}
