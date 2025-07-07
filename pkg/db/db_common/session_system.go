package db_common

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/constants/runtime"
)

// SystemClientExecutor is the executor function that is called within a transaction
// make sure that by the time the executor finishes execution, the connection is freed
// otherwise we will get a `conn is busy` error
type SystemClientExecutor func(context.Context, pgx.Tx) error

// ExecuteSystemClientCall creates a transaction and sets the application_name to the
// one used by the system client, executes the callback and sets the application name back to the client app name
func ExecuteSystemClientCall(ctx context.Context, conn *pgx.Conn, executor SystemClientExecutor) error {
	if !IsClientAppName(conn.Config().RuntimeParams[constants.RuntimeParamsKeyApplicationName]) {
		// this should NEVER happen
		return sperr.New("ExecuteSystemClientCall called with appname other than client: %s", conn.Config().RuntimeParams[constants.RuntimeParamsKeyApplicationName])
	}

	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) (e error) {
		// if the appName is the ClientAppName, we need to set it to ClientSystemAppName
		// and then revert when done
		_, err := tx.Exec(ctx, fmt.Sprintf("SET application_name TO '%s'", runtime.ClientSystemConnectionAppName))
		if err != nil {
			return sperr.WrapWithRootMessage(err, "could not set application name on connection")
		}
		defer func() {
			// set back the original application name
			_, err = tx.Exec(ctx, fmt.Sprintf("SET application_name TO '%s'", conn.Config().RuntimeParams[constants.RuntimeParamsKeyApplicationName]))
			if err != nil {
				log.Println("[TRACE] could not reset application_name", e)
			}
			// if there is not already an error, set the error
			if e == nil {
				e = err
			}
		}()

		if err := executor(ctx, tx); err != nil {
			return sperr.WrapWithMessage(err, "system client query execution failed")
		}
		return nil
	})
}
