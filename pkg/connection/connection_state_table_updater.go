package connection

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
	"github.com/turbot/steampipe/v2/pkg/introspection"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

type connectionStateTableUpdater struct {
	updates *steampipeconfig.ConnectionUpdates
	pool    *pgxpool.Pool
}

func newConnectionStateTableUpdater(updates *steampipeconfig.ConnectionUpdates, pool *pgxpool.Pool) *connectionStateTableUpdater {
	log.Println("[DEBUG] newConnectionStateTableUpdater start")
	defer log.Println("[DEBUG] newConnectionStateTableUpdater end")

	return &connectionStateTableUpdater{
		updates: updates,
		pool:    pool,
	}
}

// update connection state table to indicate the updates that will be done
func (u *connectionStateTableUpdater) start(ctx context.Context) error {
	log.Println("[DEBUG] connectionStateTableUpdater.start start")
	defer log.Println("[DEBUG] connectionStateTableUpdater.start end")

	var queries []db_common.QueryWithArgs

	// update the conection state table to set appropriate state for all connections
	// set updates to "updating"
	for name, connectionState := range u.updates.FinalConnectionState {
		// set the connection data state based on whether this connection is being created or deleted
		if _, updatingConnection := u.updates.Update[name]; updatingConnection {
			connectionState.State = constants.ConnectionStateUpdating
			connectionState.CommentsSet = false
		} else if validationError, connectionIsInvalid := u.updates.InvalidConnections[name]; connectionIsInvalid {
			// if this connection has an error, set to error
			connectionState.State = constants.ConnectionStateError
			connectionState.ConnectionError = &validationError.Message
		}
		// get the sql to update the connection state in the table to match the struct
		queries = append(queries, introspection.GetUpsertConnectionStateSql(connectionState)...)
	}
	// set deletions to "deleting"
	for name := range u.updates.Delete {
		// if we are we deleting the schema because schema_import="disabled", DO NOT set state to deleting -
		// it will be set to "disabled below
		if _, connectionDisabled := u.updates.Disabled[name]; connectionDisabled {
			continue
		}

		queries = append(queries, introspection.GetSetConnectionStateSql(name, constants.ConnectionStateDeleting)...)
	}

	// set any connections with import_schema=disabled to "disabled"
	// also build a lookup of disabled connections
	for name := range u.updates.Disabled {
		queries = append(queries, introspection.GetSetConnectionStateSql(name, constants.ConnectionStateDisabled)...)
	}
	conn, err := u.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err = db_local.ExecuteSqlWithArgsInTransaction(ctx, conn.Conn(), queries...); err != nil {
		return err
	}
	return nil
}

func (u *connectionStateTableUpdater) onConnectionReady(ctx context.Context, conn *pgx.Conn, name string) error {
	log.Println("[DEBUG] connectionStateTableUpdater.onConnectionReady start")
	defer log.Println("[DEBUG] connectionStateTableUpdater.onConnectionReady end")

	connection := u.updates.FinalConnectionState[name]
	queries := introspection.GetSetConnectionStateSql(connection.ConnectionName, constants.ConnectionStateReady)
	for _, q := range queries {
		if _, err := conn.Exec(ctx, q.Query, q.Args...); err != nil {
			return err
		}
	}
	return nil
}

func (u *connectionStateTableUpdater) onConnectionCommentsLoaded(ctx context.Context, conn *pgx.Conn, name string) error {
	log.Println("[DEBUG] connectionStateTableUpdater.onConnectionCommentsLoaded start")
	defer log.Println("[DEBUG] connectionStateTableUpdater.onConnectionCommentsLoaded end")

	connection := u.updates.FinalConnectionState[name]
	queries := introspection.GetSetConnectionStateCommentLoadedSql(connection.ConnectionName, true)
	for _, q := range queries {
		if _, err := conn.Exec(ctx, q.Query, q.Args...); err != nil {
			return err
		}
	}
	return nil
}

func (u *connectionStateTableUpdater) onConnectionDeleted(ctx context.Context, conn *pgx.Conn, name string) error {
	log.Println("[DEBUG] connectionStateTableUpdater.onConnectionDeleted start")
	defer log.Println("[DEBUG] connectionStateTableUpdater.onConnectionDeleted end")

	// if this connection has schema import disabled, DO NOT delete from the conneciotn state table
	if _, connectionDisabled := u.updates.Disabled[name]; connectionDisabled {
		return nil
	}
	queries := introspection.GetDeleteConnectionStateSql(name)
	for _, q := range queries {
		if _, err := conn.Exec(ctx, q.Query, q.Args...); err != nil {
			return err
		}
	}
	return nil
}

func (u *connectionStateTableUpdater) onConnectionError(ctx context.Context, conn *pgx.Conn, connectionName string, err error) error {
	log.Println("[DEBUG] connectionStateTableUpdater.onConnectionError start")
	defer log.Println("[DEBUG] connectionStateTableUpdater.onConnectionError end")

	queries := introspection.GetConnectionStateErrorSql(connectionName, err)
	for _, q := range queries {
		if _, err := conn.Exec(ctx, q.Query, q.Args...); err != nil {
			return err
		}
	}

	return nil
}
