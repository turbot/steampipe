package utils

import (
	"context"
	"errors"
)

// ContextCancelled is a helper function which returns whether the context has been cancelled
func ContextCancelled(ctx context.Context) bool {
	return errors.Is(ctx.Err(), context.Canceled)
}
