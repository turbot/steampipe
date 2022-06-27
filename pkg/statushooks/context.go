package statushooks

import (
	"context"

	"github.com/turbot/steampipe/pkg/contexthelpers"
)

var (
	contextKeyStatusHook = contexthelpers.ContextKey("status_hook")
)

func DisableStatusHooks(ctx context.Context) context.Context {
	return AddStatusHooksToContext(ctx, NullHooks)
}

func AddStatusHooksToContext(ctx context.Context, statusHooks StatusHooks) context.Context {
	return context.WithValue(ctx, contextKeyStatusHook, statusHooks)
}

func StatusHooksFromContext(ctx context.Context) StatusHooks {
	if ctx == nil {
		return NullHooks
	}
	if val, ok := ctx.Value(contextKeyStatusHook).(StatusHooks); ok {
		return val
	}
	// no status hook in context - return null status hook
	return NullHooks
}

func SetStatus(ctx context.Context, msg string) {
	StatusHooksFromContext(ctx).SetStatus(msg)
}

func Done(ctx context.Context) {
	StatusHooksFromContext(ctx).Done()
}

func Message(ctx context.Context, msgs ...string) {
	StatusHooksFromContext(ctx).Message(msgs...)
}
