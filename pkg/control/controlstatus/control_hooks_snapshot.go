package controlstatus

import (
	"context"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/statushooks"
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

func (c *SnapshotControlHooks) OnStart(context.Context, *ControlProgress) {
}

func (c *SnapshotControlHooks) OnControlStart(context.Context, ControlRunStatusProvider, *ControlProgress) {
}

func (c *SnapshotControlHooks) OnControlComplete(ctx context.Context, _ ControlRunStatusProvider, progress *ControlProgress) {
	statushooks.UpdateSnapshotProgress(ctx, progress.StatusSummaries.TotalCount())
}

func (c *SnapshotControlHooks) OnControlError(ctx context.Context, _ ControlRunStatusProvider, _ *ControlProgress) {
	statushooks.SnapshotError(ctx)
}

func (c *SnapshotControlHooks) OnComplete(_ context.Context, _ *ControlProgress) {
}
