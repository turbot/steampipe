package db_local

import (
	"context"

	"github.com/turbot/steampipe/pkg/constants"
)

// RefreshConnectionAndSearchPaths creates a local client and refreshed connections and search paths
func RefreshConnectionAndSearchPaths(ctx context.Context, invoker constants.Invoker) error {
	client, err := NewLocalClient(ctx, invoker)
	if err != nil {
		return err
	}
	defer client.Close(ctx)
	refreshResult := client.RefreshConnectionAndSearchPaths(ctx)
	// display any initialisation warnings
	refreshResult.ShowWarnings()

	return refreshResult.Error
}
