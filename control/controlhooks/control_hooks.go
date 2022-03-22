package controlhooks

import "context"

type ControlHooks interface {
	OnStart(context.Context, *ControlProgress)
	OnControlEvent(context.Context, *ControlProgress)
	OnDone(context.Context, *ControlProgress)
}
