package controlhooks

import "context"

type ControlHooks interface {
	OnControlEvent(context.Context, *ControlProgress)
	OnDone(context.Context, *ControlProgress)
}
