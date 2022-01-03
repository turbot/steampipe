package db_local

import (
	"context"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/statushooks"
)

// RefreshConnectionAndSearchPaths creates a local client and refreshed connections and search paths
func RefreshConnectionAndSearchPaths(ctx context.Context, invoker constants.Invoker, statushook statushooks.StatusHooks) error {
	client, err := NewLocalClient(ctx, invoker, statushook)
	if err != nil {
		return err
	}
	defer client.Close()
	refreshResult := client.RefreshConnectionAndSearchPaths(ctx)
	// display any initialisation warnings
	refreshResult.ShowWarnings()

	return refreshResult.Error
}
