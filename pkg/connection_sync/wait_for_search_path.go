package connection_sync

import (
	"context"
	"github.com/turbot/steampipe/pkg/db/steampipe_db_client"
	"github.com/turbot/steampipe/pkg/steampipe_config_local"

	"github.com/turbot/pipe-fittings/db_common"
)

// WaitForSearchPathSchemas identifies the first connection in the search path for each plugin,
// and wait for these connections to be ready
// if any of the connections are in error state, return an error
// this is used to ensure unqualified queries and tables are resolved to the correct connection
func WaitForSearchPathSchemas(ctx context.Context, client steampipe_db_client.SteampipeDbClient, searchPath []string) error {
	conn, err := client.AcquireManagementConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = steampipe_config_local.LoadConnectionState(ctx, conn, steampipe_config_local.WithWaitForSearchPath(searchPath))

	// NOTE: if we failed to load conection state, this must be because we are connected to an older version of the CLI
	// just return nil error
	if db_common.IsRelationNotFoundError(err) {
		return nil
	}

	return err
}
