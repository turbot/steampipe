package db_local

// // RefreshConnectionAndSearchPaths creates a local client and refreshed connections and search paths
// func RefreshConnectionAndSearchPaths(ctx context.Context, invoker constants.Invoker) (*steampipeconfig.RefreshConnectionResult, error) {
// 	client, err := NewLocalClient(ctx, invoker, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer client.Close(ctx)
// 	refreshResult := client.RefreshConnectionAndSearchPaths(ctx)
// 	return refreshResult, nil
// }
