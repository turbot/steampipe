package connection

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/steampipe/pkg/connection/connection_state"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"log"
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

	for name, connectionState := range u.updates.FinalConnectionState {
		// set the connection data state based on whether this connection is being created or deleted
		if _, updatingConnection := u.updates.Update[name]; updatingConnection {
			connectionState.State = constants.ConnectionStateUpdating
		}
		queries = append(queries, connection_state.GetStartUpdateConnectionStateSql(connectionState))
	}
	for name := range u.updates.Delete {
		queries = append(queries, connection_state.GetSetConnectionDeletingSql(name))
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
	q := connection_state.GetSetConnectionReadySql(connection)
	_, err := conn.Exec(ctx, q.Query, q.Args...)
	if err != nil {
		return err
	}

	return nil
}

func (u *connectionStateTableUpdater) onConnectionDeleted(ctx context.Context, conn *pgx.Conn, name string) error {
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
