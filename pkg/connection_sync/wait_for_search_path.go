package connection_sync

import (
	"context"

	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// WaitForSearchPathSchemas identifies the first connection in the search path for each plugin,
// and wait for these connections to be ready
// if any of the connections are in error state, return an error
// this is used to ensure unqualified queries and tables are resolved to the correct connection
func WaitForSearchPathSchemas(ctx context.Context, client db_common.Client, searchPath []string) error {
	conn, err := client.AcquireManagementConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = steampipeconfig.LoadConnectionState(ctx, conn.Conn(), steampipeconfig.WithWaitForSearchPath(searchPath))

	// NOTE: if we failed to load conection state, this must be because we are connected to an older version of the CLI
	// just return nil error
	if db_common.IsRelationNotFoundError(err) {
		return nil
	}

	return err
}
