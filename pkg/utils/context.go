package utils

import (
	"context"
	"github.com/turbot/steampipe-plugin-sdk/v4/error_helpers"
)

func IsContextCancelled(ctx context.Context) bool {
	return error_helpers.IsContextCancelledError(ctx.Err())
}
