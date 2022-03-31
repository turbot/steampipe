package controlstatus

import (
	"context"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ControlHooks interface {
	OnStart(context.Context, *ControlProgress)
	OnControlStart(context.Context, *modconfig.Control, *ControlProgress)
	OnControlComplete(context.Context, *modconfig.Control, ControlRunStatus, *StatusSummary, *ControlProgress)
	OnControlError(context.Context, *modconfig.Control, ControlRunStatus, *StatusSummary, *ControlProgress)
	OnComplete(context.Context, *ControlProgress)
}
