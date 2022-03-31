package controlstatus

import (
	"context"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

var NullHooks = &NullControlHook{}

type NullControlHook struct{}

func (*NullControlHook) OnStart(context.Context, *ControlProgress) {
}
func (*NullControlHook) OnControlStart(context.Context, *modconfig.Control, *ControlProgress) {
}
func (*NullControlHook) OnControlComplete(context.Context, *modconfig.Control, ControlRunStatus, *StatusSummary, *ControlProgress) {
}
func (*NullControlHook) OnControlError(context.Context, *modconfig.Control, ControlRunStatus, *StatusSummary, *ControlProgress) {
}
func (*NullControlHook) OnComplete(context.Context, *ControlProgress) {}
