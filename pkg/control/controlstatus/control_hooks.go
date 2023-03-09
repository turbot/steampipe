package controlstatus

import (
	"context"
)

type ControlHooks interface {
	OnStart(context.Context, *ControlProgress)
	OnControlStart(context.Context, ControlRunStatusProvider, *ControlProgress)
	OnControlComplete(context.Context, ControlRunStatusProvider, *ControlProgress)
	OnControlError(context.Context, ControlRunStatusProvider, *ControlProgress)
	OnComplete(context.Context, *ControlProgress)
}
