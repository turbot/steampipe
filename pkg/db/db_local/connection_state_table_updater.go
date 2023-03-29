package db_local

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/constants"
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
	var queries []string
	for c := range u.updates.RequiredConnectionState {
		_, updateConnection := u.updates.Update[c]
		_, deleteConnection := u.updates.Update[c]
		if !updateConnection && !deleteConnection {
			queries = append(queries, getUpdateConnectionStateSql(c, constants.ConnectionStateReady))
		}
	}
	for c := range u.updates.Update {
		queries = append(queries, getUpdateConnectionStateSql(c, constants.ConnectionStateUpdating))
	}
	for c := range u.updates.Delete {
		queries = append(queries, getUpdateConnectionStateSql(c, constants.ConnectionStateDeleting))
	}
	if _, err := executeSqlAsRoot(ctx, queries...); err != nil {
		return err
	}
	return nil
}

func (u *connectionStateTableUpdater) onConnectionUpdated(ctx context.Context, tx pgx.Tx, name string) error {
	sql := getUpdateConnectionStateSql(name, constants.ConnectionStateReady)
	_, err := tx.Exec(ctx, sql)
	if err != nil {
		return err
	}

	// this may be called from multiple goroutines
	u.updateCountLock.Lock()
	u.updatedConnections++
	if u.updatedConnections >= updatePageSize {
		log.Printf("[WARN] onConnectionUpdated updating page")
		// TODO KAI send notification
		//statushooks.SetStatus(ctx, fmt.Sprintf("Cloned %d of %d %s (%s)", idx, numUpdates, utils.Pluralize("connection", numUpdates), connectionName))
		u.updatedConnections = 0
	}
	u.updateCountLock.Unlock()

	return nil
}

func (u *connectionStateTableUpdater) onConnectionDeleted(ctx context.Context, tx pgx.Tx, name string) error {
	sql := getDeleteConnectionStateSql(name)
	_, err := tx.Exec(ctx, sql)
	if err != nil {
		return err
	}
	// this may be called from multiple goroutines
	u.deleteCountLock.Lock()

	u.deletedConnections++
	if u.deletedConnections >= updatePageSize {
		log.Printf("[WARN] onConnectionDeleted updating page")
		// TODO KAI send notification

		u.deletedConnections = 0
	}
	u.deleteCountLock.Unlock()

	return nil
}

func (u *connectionStateTableUpdater) onConnectionError(ctx context.Context, tx pgx.Tx, connectionName string, err error) error {
	sql := getConnectionStateErrorSql(connectionName, err)
	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}
	return nil
	// TODO KAI send notification
}

func getConnectionStateErrorSql(connectionName string, err error) string {

	return fmt.Sprintf(`UPDATE %s.%s
SET status = '%s',
	details = '%s',
	last_change = now()
WHERE
	name '%s'
	`,

		constants.InternalSchema, constants.ConnectionStateTable,
		constants.ConnectionStateError,
		err.Error(),
		connectionName)
}

func getUpdateConnectionStateSql(connectionName, state string) string {
	// upsert
	return fmt.Sprintf(`INSERT INTO %s.%s (name, status, last_change)
VALUES('%s','%s', now()) 
ON CONFLICT (name) 
DO 
   UPDATE SET status = '%s', 
			  details = null,
			  comments_set = false,
			  last_change = now()
`,
		constants.InternalSchema, constants.ConnectionStateTable,
		connectionName, state,
		state)
}
func getDeleteConnectionStateSql(connectionName string) string {
	return fmt.Sprintf(`DELETE FROM %s.%s WHERE NAME='%s'`, connectionName)
}
