package statushooks

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/pkg/contexthelpers"
)

var (
	contextKeySnapshotProgress = contexthelpers.ContextKey("snapshot_progress")
	contextKeyStatusHook       = contexthelpers.ContextKey("status_hook")
	contextKeyMessageRenderer  = contexthelpers.ContextKey("message_renderer")
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

func AddSnapshotProgressToContext(ctx context.Context, snapshotProgress SnapshotProgress) context.Context {
	return context.WithValue(ctx, contextKeySnapshotProgress, snapshotProgress)
}

func SnapshotProgressFromContext(ctx context.Context) SnapshotProgress {
	if ctx == nil {
		return NullProgress
	}
	if val, ok := ctx.Value(contextKeySnapshotProgress).(SnapshotProgress); ok {
		return val
	}
	// no snapshot progress in context - return null progress
	return NullProgress
}

func AddMessageRendererToContext(ctx context.Context, messageRenderer MessageRenderer) context.Context {
	return context.WithValue(ctx, contextKeyMessageRenderer, messageRenderer)
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
