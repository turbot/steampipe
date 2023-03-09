package error_helpers

import (
	"context"
	sdkerrorhelpers "github.com/turbot/steampipe-plugin-sdk/v5/error_helpers"
)

func IsContextCanceled(ctx context.Context) bool {
	return sdkerrorhelpers.IsContextCancelledError(ctx.Err())
}

func IsContextCancelledError(err error) bool {
	return sdkerrorhelpers.IsContextCancelledError(err)
}
