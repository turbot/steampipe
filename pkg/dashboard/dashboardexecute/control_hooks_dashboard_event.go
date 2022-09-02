package dashboardexecute

import (
	"context"
	"github.com/turbot/steampipe/pkg/control/controlstatus"

	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
)

// DashboardEventControlHooks is a struct which implements ControlHooks,
// and raises ControlComplete and ControlError dashboard events
type DashboardEventControlHooks struct {
	CheckRun *CheckRun
}

func NewDashboardEventControlHooks(r *CheckRun) *DashboardEventControlHooks {
	return &DashboardEventControlHooks{
		CheckRun: r,
	}
}

func (c *DashboardEventControlHooks) OnStart(ctx context.Context, _ *controlstatus.ControlProgress) {
	// nothing to do
}

func (c *DashboardEventControlHooks) OnControlStart(context.Context, controlstatus.ControlRunStatusProvider, *controlstatus.ControlProgress) {
}

func (c *DashboardEventControlHooks) OnControlComplete(ctx context.Context, controlRun controlstatus.ControlRunStatusProvider, progress *controlstatus.ControlProgress) {
	event := &dashboardevents.ControlComplete{
		Control:     controlRun,
		Progress:    progress,
		Name:        c.CheckRun.Name,
		ExecutionId: c.CheckRun.executionTree.id,
		Session:     c.CheckRun.SessionId,
	}
	c.CheckRun.executionTree.workspace.PublishDashboardEvent(event)
}

func (c *DashboardEventControlHooks) OnControlError(ctx context.Context, controlRun controlstatus.ControlRunStatusProvider, progress *controlstatus.ControlProgress) {
	var event = &dashboardevents.ControlError{
		Control:     controlRun,
		Progress:    progress,
		Name:        c.CheckRun.Name,
		ExecutionId: c.CheckRun.executionTree.id,
		Session:     c.CheckRun.SessionId,
	}
	c.CheckRun.executionTree.workspace.PublishDashboardEvent(event)
}

func (c *DashboardEventControlHooks) OnComplete(ctx context.Context, _ *controlstatus.ControlProgress) {
	// nothing to do - LeafNodeDone will be sent anyway
}
