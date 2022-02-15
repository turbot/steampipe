package dashboardexecute

import (
	"context"

	"github.com/turbot/steampipe/control/controlhooks"
	"github.com/turbot/steampipe/dashboard/dashboardevents"
)

// ControlEventHooks is a struct which implements ControlHooks, and displays the control progress as a status message
type ControlEventHooks struct {
	CheckRun *CheckRun
}

func NewControlEventHooks(r *CheckRun) *ControlEventHooks {
	return &ControlEventHooks{
		CheckRun: r,
	}
}

func (c *ControlEventHooks) OnStart(ctx context.Context, _ *controlhooks.ControlProgress) {
	// nothing to do
}

func (c *ControlEventHooks) OnControlEvent(ctx context.Context, _ *controlhooks.ControlProgress) {
	event := &dashboardevents.LeafNodeProgress{Node: c.CheckRun}
	c.CheckRun.executionTree.workspace.PublishDashboardEvent(event)
}

func (c *ControlEventHooks) OnDone(ctx context.Context, _ *controlhooks.ControlProgress) {
	// nothing to do - LeadNodeDone will be sent anyway
}
