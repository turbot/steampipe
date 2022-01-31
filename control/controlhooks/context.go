package controlhooks

import (
	"context"

	"github.com/turbot/steampipe/contexthelpers"
)

var (
	contextKeyControlHook = contexthelpers.ContextKey("control_hook")
)

func AddControlHooksToContext(ctx context.Context, statusHooks ControlHooks) context.Context {
	return context.WithValue(ctx, contextKeyControlHook, statusHooks)
}

func ControlHooksFromContext(ctx context.Context) ControlHooks {
	if ctx == nil {
		return NullHooks
	}
	if val, ok := ctx.Value(contextKeyControlHook).(ControlHooks); ok {
		return val
	}
	// no status hook in context - return null status hook
	return NullHooks
}

func OnStart(ctx context.Context, p *ControlProgress) {
	ControlHooksFromContext(ctx).OnStart(ctx, p)
}

func OnControlEvent(ctx context.Context, p *ControlProgress) {
	ControlHooksFromContext(ctx).OnControlEvent(ctx, p)
}

func OnDone(ctx context.Context, p *ControlProgress) {
	ControlHooksFromContext(ctx).OnDone(ctx, p)
}
