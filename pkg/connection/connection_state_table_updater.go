package connection

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/steampipe/pkg/connection_state"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

type connectionStateTableUpdater struct {
	updates *steampipeconfig.ConnectionUpdates
	pool    *pgxpool.Pool
}

func newConnectionStateTableUpdater(updates *steampipeconfig.ConnectionUpdates, pool *pgxpool.Pool) *connectionStateTableUpdater {
	return &connectionStateTableUpdater{
		updates: updates,
		pool:    pool,
	}
}

// update connection state table to indicate the updates that will be done
func (u *connectionStateTableUpdater) start(ctx context.Context) error {
	log.Printf("[INFO] connectionStateTableUpdater start - update connection_state with intended states")

	var queries []db_common.QueryWithArgs

	// update the conection state table to set appropriate state for all connections
	// set updates to "updating"
	for name, connectionState := range u.updates.FinalConnectionState {
		log.Printf("[INFO] >> name: %s connectionState: %s modtime: %v", name, connectionState.State, connectionState.ConnectionModTime)
		// set the connection data state based on whether this connection is being created or deleted
		if _, updatingConnection := u.updates.Update[name]; updatingConnection {
			connectionState.State = constants.ConnectionStateUpdating
			connectionState.CommentsSet = false
			log.Printf("[INFO] >> (in if) name: %s connectionState: %s modtime: %v", name, connectionState.State, connectionState.ConnectionModTime)
		} else if validationError, connectionIsInvalid := u.updates.InvalidConnections[name]; connectionIsInvalid {
			// if this connection has an error, set to error
			connectionState.State = constants.ConnectionStateError
			connectionState.ConnectionError = &validationError.Message
			log.Printf("[INFO] >> (in else) name: %s connectionState: %s modtime: %v", name, connectionState.State, connectionState.ConnectionModTime)
		}
		// get the sql to update the connection state in the table to match the struct
		queries = append(queries, connection_state.GetUpdateConnectionStateSql(connectionState))
	}
	// set deletions to "deleting"
	for name := range u.updates.Delete {
		log.Printf("[INFO] >>> name: %s", name)
		// if we are we deleting the schema because schema_import="disabled", DO NOT set state to deleting -
		// it will be set to "disabled below
		if _, connectionDisabled := u.updates.Disabled[name]; connectionDisabled {
			continue
		}

		queries = append(queries, connection_state.GetSetConnectionStateSql(name, constants.ConnectionStateDeleting))
	}

	// set any connections with import_schema=disabled to "disabled"
	// also build a lookup of disabled connections
	for name := range u.updates.Disabled {
		queries = append(queries, connection_state.GetSetConnectionStateSql(name, constants.ConnectionStateDisabled))
	}
	conn, err := u.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err = db_local.ExecuteSqlWithArgsInTransaction(ctx, conn.Conn(), queries...); err != nil {
		return err
	}
	log.Printf("[INFO] connectionStateTableUpdater start - finished updating connection_state with intended states")
	return nil
}

func (u *connectionStateTableUpdater) onConnectionReady(ctx context.Context, conn *pgx.Conn, name string) error {
	connection := u.updates.FinalConnectionState[name]
	q := connection_state.GetSetConnectionStateSql(connection.ConnectionName, constants.ConnectionStateReady)
	log.Printf("[INFO] >> connection %v", connection.ConnectionModTime)
	_, err := conn.Exec(ctx, q.Query, q.Args...)
	if err != nil {
		return err
	}

	return nil
}

func (u *connectionStateTableUpdater) onConnectionCommentsLoaded(ctx context.Context, conn *pgx.Conn, name string) error {
	connection := u.updates.FinalConnectionState[name]
	q := connection_state.GetSetConnectionStateCommentLoadedSql(connection.ConnectionName, true)
	_, err := conn.Exec(ctx, q.Query, q.Args...)
	if err != nil {
		return err
	}

	return nil
}

func (u *connectionStateTableUpdater) onConnectionDeleted(ctx context.Context, conn *pgx.Conn, name string) error {
	// if this connection has schema import disabled, DO NOT delete from the conneciotn state table
	if _, connectionDisabled := u.updates.Disabled[name]; connectionDisabled {
		return nil
	}
	q := connection_state.GetDeleteConnectionStateSql(name)
	_, err := conn.Exec(ctx, q.Query, q.Args...)
	if err != nil {
		return err
	}

	return nil
}

func (u *connectionStateTableUpdater) onConnectionError(ctx context.Context, conn *pgx.Conn, connectionName string, err error) error {
	q := connection_state.GetConnectionStateErrorSql(connectionName, err)
	if _, err := conn.Exec(ctx, q.Query, q.Args...); err != nil {
		return err
	}

	return nil
	// TODO send notification
}
