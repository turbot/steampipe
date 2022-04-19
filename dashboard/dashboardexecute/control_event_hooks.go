package dashboardexecute

import (
	"context"

	"github.com/turbot/steampipe/control/controlstatus"
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

func (c *ControlEventHooks) OnStart(ctx context.Context, _ *controlstatus.ControlProgress) {
	// nothing to do
}

func (c *ControlEventHooks) OnControlStart(context.Context, controlstatus.ControlRunStatusProvider, *controlstatus.ControlProgress) {
}

func (c *ControlEventHooks) OnControlComplete(ctx context.Context, controlRun controlstatus.ControlRunStatusProvider, progress *controlstatus.ControlProgress) {
	event := &dashboardevents.ControlComplete{
		Control:     controlRun,
		Progress:    progress,
		Name:        c.CheckRun.Name,
		ExecutionId: c.CheckRun.executionTree.id,
		Session:     c.CheckRun.SessionId,
	}
	c.CheckRun.executionTree.workspace.PublishDashboardEvent(event)
}

func (c *ControlEventHooks) OnControlError(ctx context.Context, controlRun controlstatus.ControlRunStatusProvider, progress *controlstatus.ControlProgress) {
	var event = &dashboardevents.ControlError{
		Control:     controlRun,
		Progress:    progress,
		Name:        c.CheckRun.Name,
		ExecutionId: c.CheckRun.executionTree.id,
		Session:     c.CheckRun.SessionId,
	}
	c.CheckRun.executionTree.workspace.PublishDashboardEvent(event)
}

func (c *ControlEventHooks) OnComplete(ctx context.Context, _ *controlstatus.ControlProgress) {
	// nothing to do - LeafNodeDone will be sent anyway
}
