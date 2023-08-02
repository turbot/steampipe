package db_common

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/constants/runtime"
)

type SystemClientExecutor func(context.Context, pgx.Tx) error

// ExecuteSystemClientCallOnConnection creates a transaction and sets the application_name to the
// one used by the system client, executes the callback and sets the application name back to the client app name
func ExecuteSystemClientCallOnConnection(ctx context.Context, conn *pgx.Conn, executor SystemClientExecutor) error {
	// TODO:: should we check the application name first so that we don't set it back to something incorrect?
	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		_, err := conn.Exec(ctx, "SET application_name TO $1", runtime.ClientSystemConnectionAppName)
		if err != nil {
			return err
		}
		if err := executor(ctx, tx); err != nil {
			return err
		}
		_, err = conn.Exec(ctx, "SET application_name TO $1", runtime.ClientConnectionAppName)
		if err != nil {
			return err
		}
		return nil
	})
}
