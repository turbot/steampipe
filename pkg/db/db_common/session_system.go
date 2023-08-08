package db_common

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/constants/runtime"
)

// SystemClientExecutor is the executor function that is called within a transaction
// make sure that by the time the executor finishes execution, the connection is freed
// otherwise we will get a `conn is busy` error
type SystemClientExecutor func(context.Context, pgx.Tx) error

// ExecuteSystemClientCall creates a transaction and sets the application_name to the
// one used by the system client, executes the callback and sets the application name back to the client app name
func ExecuteSystemClientCall(ctx context.Context, conn *pgx.Conn, executor SystemClientExecutor) error {
	// checks that the appname is the one reserved for user-originating queries
	// we need to check this since we may be calling this function with connections created
	// from the system pool as well - in which case, we will not need to update the app name
	appNameNeedsUpdate := IsClientAppName(conn.Config().RuntimeParams[constants.RuntimeParamsKeyApplicationName])

	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) (e error) {
		// if the appName is the ClientAppName, we need to set it to ClientSystemAppName
		// and then revert when done
		if appNameNeedsUpdate {
			_, err := tx.Exec(ctx, fmt.Sprintf("SET application_name TO '%s'", runtime.ClientSystemConnectionAppName))
			if err != nil {
				return err
			}
			defer func() {
				// set back the original application name
				_, e = tx.Exec(ctx, fmt.Sprintf("SET application_name TO '%s'", conn.Config().RuntimeParams[constants.RuntimeParamsKeyApplicationName]))
			}()
		}

		if err := executor(ctx, tx); err != nil {
			return err
		}
		return nil
	})
}
