package local_db

// RefreshConnectionAndSearchPaths creates a local client and refreshed connections and search paths
func RefreshConnectionAndSearchPaths(client *LocalClient) error {
	refreshResult := client.RefreshConnectionAndSearchPaths()
	// display any initialisation warnings
	refreshResult.ShowWarnings()

	return refreshResult.Error
}
