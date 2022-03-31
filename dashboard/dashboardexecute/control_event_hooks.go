package dashboardexecute

import (
	"context"

	"github.com/turbot/steampipe/control/controlstatus"
	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
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

func (c *ControlEventHooks) OnControlStart(ctx context.Context, control *modconfig.Control, progress *controlstatus.ControlProgress) {
	event := &dashboardevents.LeafNodeProgress{
		LeafNode:    c.CheckRun,
		ExecutionId: c.CheckRun.executionTree.id,
		Session:     c.CheckRun.SessionId,
	}
	c.CheckRun.executionTree.workspace.PublishDashboardEvent(event)
}

func (c *ControlEventHooks) OnControlComplete(ctx context.Context, control *modconfig.Control, controlRunStatus controlstatus.ControlRunStatus, controlStatusSummary *controlstatus.StatusSummary, progress *controlstatus.ControlProgress) {
	event := &dashboardevents.ControlComplete{
		ControlName:          control.Name(),
		ControlRunStatus:     controlRunStatus,
		ControlStatusSummary: controlStatusSummary,
		Progress:             progress,
		ExecutionId:          c.CheckRun.executionTree.id,
		Session:              c.CheckRun.SessionId,
	}
	c.CheckRun.executionTree.workspace.PublishDashboardEvent(event)
}

func (c *ControlEventHooks) OnControlError(ctx context.Context, control *modconfig.Control, controlRunStatus controlstatus.ControlRunStatus, controlStatusSummary *controlstatus.StatusSummary, progress *controlstatus.ControlProgress) {
	var event = &dashboardevents.ControlError{
		ControlName:          control.Name(),
		ControlRunStatus:     controlRunStatus,
		ControlStatusSummary: controlStatusSummary,
		Progress:             progress,
		ExecutionId:          c.CheckRun.executionTree.id,
		Session:              c.CheckRun.SessionId,
	}
	c.CheckRun.executionTree.workspace.PublishDashboardEvent(event)
}

func (c *ControlEventHooks) OnComplete(ctx context.Context, _ *controlstatus.ControlProgress) {
	// nothing to do - LeafNodeDone will be sent anyway
}
