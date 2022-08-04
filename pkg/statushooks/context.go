package statushooks

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/pkg/contexthelpers"
)

var (
	contextKeyStatusHook      = contexthelpers.ContextKey("status_hook")
	contextKeyMessageRenderer = contexthelpers.ContextKey("meddage_renderer")
)

func DisableStatusHooks(ctx context.Context) context.Context {
	return AddStatusHooksToContext(ctx, NullHooks)
}

func AddStatusHooksToContext(ctx context.Context, statusHooks StatusHooks) context.Context {
	return context.WithValue(ctx, contextKeyStatusHook, statusHooks)
}

func AddMessageRendererToContext(ctx context.Context, messageRenderer MessageRenderer) context.Context {
	return context.WithValue(ctx, contextKeyMessageRenderer, messageRenderer)
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

type MessageRenderer func(format string, a ...any)

func MessageRendererFromContext(ctx context.Context) MessageRenderer {
	defaultRenderer := func(format string, a ...any) {
		fmt.Printf(format, a...)
	}
	if ctx == nil {
		return defaultRenderer
	}
	if val, ok := ctx.Value(contextKeyMessageRenderer).(MessageRenderer); ok {
		return val
	}
	// no message renderer - return fmt.Printf
	return defaultRenderer
}
