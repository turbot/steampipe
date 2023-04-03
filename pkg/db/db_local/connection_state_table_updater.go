package db_local

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"log"
	"sync"
)

// send a notification every 5 connections
const updatePageSize = 5

type connectionStateTableUpdater struct {
	updates *steampipeconfig.ConnectionUpdates

	updateCountLock    sync.Mutex
	deleteCountLock    sync.Mutex
	updatedConnections int
	deletedConnections int
}

func newConnectionStateTableUpdater(updates *steampipeconfig.ConnectionUpdates) *connectionStateTableUpdater {
	return &connectionStateTableUpdater{
		updates: updates,
	}
}

// update connection state table to indicate the updates that will be done
func (u *connectionStateTableUpdater) start(ctx context.Context) error {
	log.Printf("[WARN] connectionStateTableUpdater start")

	var queries []db_common.QueryWithArgs

	for name, connectionData := range u.updates.FinalConnectionState {
		// set the connection data state based on whether this connection is being created or deleted
		if _, updatingConnection := u.updates.Update[name]; updatingConnection {
			connectionData.ConnectionState = constants.ConnectionStateUpdating
		} else if _, deletingConnection := u.updates.Update[name]; deletingConnection {
			connectionData.ConnectionState = constants.ConnectionStateDeleting
		}
		queries = append(queries, getStartUpdateConnectionStateSql(connectionData))
	}

	if _, err := executeSqlWithArgsAsRoot(ctx, queries...); err != nil {
		return err
	}
	return nil
}

func (u *connectionStateTableUpdater) onConnectionReady(ctx context.Context, tx pgx.Tx, name string) error {
	connection := u.updates.FinalConnectionState[name]
	q := getConnectionReadySql(connection)
	_, err := tx.Exec(ctx, q.Query, q.Args...)
	if err != nil {
		return err
	}

	// this may be called from multiple goroutines
	u.updateCountLock.Lock()
	u.updatedConnections++
	if u.updatedConnections >= updatePageSize {
		//log.Printf("[WARN] onConnectionReady updating page")
		// TODO KAI send notification
		//statushooks.SetStatus(ctx, fmt.Sprintf("Cloned %d of %d %s (%s)", idx, numUpdates, utils.Pluralize("connection", numUpdates), connectionName))
		u.updatedConnections = 0
	}
	u.updateCountLock.Unlock()

	return nil
}

func (u *connectionStateTableUpdater) onConnectionDeleted(ctx context.Context, tx pgx.Tx, name string) error {
	q := getDeleteConnectionStateSql(name)
	_, err := tx.Exec(ctx, q.Query, q.Args...)
	if err != nil {
		return err
	}
	// this may be called from multiple goroutines
	u.deleteCountLock.Lock()

	u.deletedConnections++
	if u.deletedConnections >= updatePageSize {
		//log.Printf("[WARN] onConnectionDeleted updating page")
		// TODO KAI send notification

		u.deletedConnections = 0
	}
	u.deleteCountLock.Unlock()

	return nil
}

func (u *connectionStateTableUpdater) onConnectionError(ctx context.Context, tx pgx.Tx, connectionName string, err error) error {
	q := getConnectionStateErrorSql(connectionName, err)
	if _, err := tx.Exec(ctx, q.Query, q.Args...); err != nil {
		return err
	}
	return nil
	// TODO KAI send notification
}

func getConnectionStateErrorSql(connectionName string, err error) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE %s.%s
SET state = $1,
	error = $2,
	connection_mod_time = now()
WHERE
	name $3
	`,
		constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{constants.ConnectionStateError, err.Error(), connectionName}
	return db_common.QueryWithArgs{query, args}
}

func getStartUpdateConnectionStateSql(c *steampipeconfig.ConnectionData) db_common.QueryWithArgs {
	// if state is updating, set comments to false
	commentsSet := c.ConnectionState == constants.ConnectionStateReady
	// upsert
	query := fmt.Sprintf(`INSERT INTO %s.%s (name, 
		state,
		error,
		plugin,
		schema_mode,
		schema_hash,
		comments_set,
		connection_mod_time,
		plugin_mod_time)
VALUES($1,$2,$3,$4,$5,$6,$7,now(),$8) 
ON CONFLICT (name) 
DO 
   UPDATE SET 
 			  state = $2, 
			  error = $3,
			  plugin = $4,
			  schema_mode = $5,
			  schema_hash = $6,
			  comments_set = $7,
			  connection_mod_time = now(),
			  plugin_mod_time = $8
`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{c.Connection.Name, c.ConnectionState, c.ConnectionError, c.Plugin, c.SchemaMode, c.SchemaHash, commentsSet, c.PluginModTime}
	return db_common.QueryWithArgs{query, args}
}

// note: set comments to false as this is called from start and updateConnectionState - both before comment completion
func getConnectionReadySql(connection *steampipeconfig.ConnectionData) db_common.QueryWithArgs {
	// upsert
	query := fmt.Sprintf(`UPDATE %s.%s 
    SET	state = $1, 
	 	connection_mod_time = now(),
	 	plugin_mod_time = $2
    WHERE 
        name = $3
`,
		constants.InternalSchema, constants.ConnectionStateTable,
	)
	args := []any{constants.ConnectionStateReady, connection.PluginModTime, connection.ConnectionName}
	return db_common.QueryWithArgs{query, args}
}

func getDeleteConnectionStateSql(connectionName string) db_common.QueryWithArgs {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE NAME=$1`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{connectionName}
	return db_common.QueryWithArgs{query, args}
}

func getSetConnectionStateCommentLoadedSql(connectionName string, commentsLoaded bool) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE  %s.%s
SET comments_loaded = $1
WHERE NAME=$2`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{commentsLoaded, connectionName}
	return db_common.QueryWithArgs{query, args}
}
