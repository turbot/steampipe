package db_common

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/constants/runtime"
)

// SystemClientExecutor is the executor function that is called within a transaction
// make sure that by the time the executor finishes execution, the connection is freed
// otherwise we will get a `conn is busy` error
type SystemClientExecutor func(context.Context, *sql.Tx) error

// ExecuteSystemClientCall creates a transaction and sets the application_name to the
// one used by the system client, executes the callback and sets the application name back to the client app name
func ExecuteSystemClientCall(ctx context.Context, conn *sql.Conn, executor SystemClientExecutor) error {
	return BeginFunc(ctx, conn, func(tx *sql.Tx) (e error) {
		// if the appName is the ClientAppName, we need to set it to ClientSystemAppName
		// and then revert when done
		_, err := tx.ExecContext(ctx, fmt.Sprintf("SET application_name TO '%s'", runtime.ClientSystemConnectionAppName))
		if err != nil {
			return sperr.WrapWithRootMessage(err, "could not set application name on connection")
		}
		defer func() {
			// set back the original application name
			_, e = tx.ExecContext(ctx, fmt.Sprintf("SET application_name TO '%s'", runtime.ClientConnectionAppName))
			if e != nil {
				log.Println("[TRACE] could not reset application_name", e)
			}
		}()

		if err := executor(ctx, tx); err != nil {
			return sperr.WrapWithMessage(err, "scoped execution failed with management client")
		}
		return nil
	})
}

type TxBeginner interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

func BeginFunc(ctx context.Context, db TxBeginner, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	return beginFuncExec(ctx, tx, fn)
}

func beginFuncExec(ctx context.Context, tx *sql.Tx, fn func(*sql.Tx) error) (err error) {
	defer func() {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = rollbackErr
		}
	}()

	fErr := fn(tx)
	if fErr != nil {
		_ = tx.Rollback() // ignore rollback error as there is already an error to return
		return fErr
	}

	return tx.Commit()
}
