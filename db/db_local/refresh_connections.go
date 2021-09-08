package db_local

import (
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
)

// RefreshConnectionAndSearchPaths creates a local client and refreshed connections and search paths
func RefreshConnectionAndSearchPaths(invoker constants.Invoker) error {
	var refreshResult = &db_common.RefreshConnectionResult{}
	client, err := NewLocalClient(invoker)
	if err != nil {
		return err
	}
	defer client.Close()
	refreshResult = client.RefreshConnectionAndSearchPaths()
	// display any initialisation warnings
	refreshResult.ShowWarnings()

	return refreshResult.Error
}
