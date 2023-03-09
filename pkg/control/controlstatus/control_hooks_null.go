package controlstatus

import (
	"context"
)

var NullHooks = &NullControlHook{}

type NullControlHook struct{}

func (*NullControlHook) OnStart(context.Context, *ControlProgress) {
}
func (*NullControlHook) OnControlStart(context.Context, ControlRunStatusProvider, *ControlProgress) {
}
func (*NullControlHook) OnControlComplete(context.Context, ControlRunStatusProvider, *ControlProgress) {
}
func (*NullControlHook) OnControlError(context.Context, ControlRunStatusProvider, *ControlProgress) {
}
func (*NullControlHook) OnComplete(context.Context, *ControlProgress) {}
