package controlstatus

import (
	"context"

	"github.com/turbot/steampipe/pkg/contexthelpers"
)

var (
	contextKeyControlHook = contexthelpers.ContextKey("control_hook")
)

func AddControlHooksToContext(ctx context.Context, statusHooks ControlHooks) context.Context {
	// if the context already contains ControlHooks, do nothing
	// this may happen when executing a dashboard snapshot -
	if _, ok := ctx.Value(contextKeyControlHook).(ControlHooks); ok {
		return ctx
	}

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

func OnControlStart(ctx context.Context, controlRun ControlRunStatusProvider, p *ControlProgress) {
	ControlHooksFromContext(ctx).OnControlStart(ctx, controlRun, p)
}

func OnControlComplete(ctx context.Context, controlRun ControlRunStatusProvider, p *ControlProgress) {
	ControlHooksFromContext(ctx).OnControlComplete(ctx, controlRun, p)
}

func OnControlError(ctx context.Context, controlRun ControlRunStatusProvider, p *ControlProgress) {
	ControlHooksFromContext(ctx).OnControlError(ctx, controlRun, p)
}

func OnComplete(ctx context.Context, p *ControlProgress) {
	ControlHooksFromContext(ctx).OnComplete(ctx, p)
}
