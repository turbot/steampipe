package db_common

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/constants/runtime"
)

// SystemClientExecutor is the executor function that is called within a transaction
// make sure that by the time the executor finishes execution, the connection is freed
// otherwise we will get a `conn is busy` error
type SystemClientExecutor func(context.Context, pgx.Tx) error

// ExecuteSystemClientCallOnConnection creates a transaction and sets the application_name to the
// one used by the system client, executes the callback and sets the application name back to the client app name
func ExecuteSystemClientCallOnConnection(ctx context.Context, conn *pgx.Conn, executor SystemClientExecutor) error {
	// checks that the appname is the one reserved for user-originating queries
	appNameNeedsUpdate := IsClientAppName(conn.Config().RuntimeParams["application_name"])

	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) (e error) {
		// if the appName is the ClientAppName, we need to set it to ClientSystemAppName
		// and then revert when done
		if appNameNeedsUpdate {
			_, err := tx.Exec(ctx, fmt.Sprintf("SET application_name TO '%s'", runtime.ClientSystemConnectionAppName))
			if err != nil {
				return err
			}
			defer func() {
				_, e = tx.Exec(ctx, fmt.Sprintf("SET application_name TO '%s'", conn.Config().RuntimeParams["application_name"]))
			}()
		}

		if err := executor(ctx, tx); err != nil {
			return err
		}
		return nil
	})
}
