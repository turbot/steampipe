package db_local

import (
	"context"
	"github.com/turbot/steampipe/pkg/steampipeconfig"

	"github.com/turbot/steampipe/pkg/constants"
)

// RefreshConnectionAndSearchPathsWithLocalClient creates a local client and refreshed connections and search paths
func RefreshConnectionAndSearchPathsWithLocalClient(ctx context.Context, invoker constants.Invoker) *steampipeconfig.RefreshConnectionResult {
	client, err := NewLocalClient(ctx, invoker, nil)
	if err != nil {
		return steampipeconfig.NewErrorRefreshConnectionResult(err)
	}
	defer client.Close(ctx)
	refreshResult := RefreshConnectionAndSearchPaths(ctx, client)
	return refreshResult
}
