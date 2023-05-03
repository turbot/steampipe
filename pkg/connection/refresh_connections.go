package connection

import (
	"context"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"log"
)

func RefreshConnectionAndSearchPaths(ctx context.Context, forceUpdateConnectionNames ...string) *steampipeconfig.RefreshConnectionResult {
	// uncomment to debug
	//time.Sleep(10 * time.Second)

	// now refresh connections
	log.Printf("[TRACE] Refreshing connections")
	// package up all necessary data into a state object
	state, err := newRefreshConnectionState(ctx, forceUpdateConnectionNames)
	if err != nil {
		return steampipeconfig.NewErrorRefreshConnectionResult(err)
	}
	defer state.close()

	state.refreshConnections(ctx)

	return state.res
}
