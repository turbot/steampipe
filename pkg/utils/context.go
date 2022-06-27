package utils

import (
	"context"
	"errors"
)

func IsContextCancelled(ctx context.Context) bool {
	err := ctx.Err()
	return err != nil && errors.Is(err, context.Canceled)
}
