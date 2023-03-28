package db_local

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"log"
)

// update the table every 5 connections
const updatePageSize = 5

type connectionStateTableUpdater struct {
	numUpdates int
	updates    *steampipeconfig.ConnectionUpdates

	updatedConnections []string
	deletedConnections []string
}

func newConnectionStateTableUpdater(updates *steampipeconfig.ConnectionUpdates) *connectionStateTableUpdater {
	return &connectionStateTableUpdater{
		numUpdates: len(updates.Update),
		updates:    updates,
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

func (u *connectionStateTableUpdater) onConnectionUpdated(ctx context.Context, name string) error {
	u.updatedConnections = append(u.updatedConnections, name)
	if len(u.updatedConnections) >= updatePageSize {
		log.Printf("[WARN] onConnectionUpdated updating page")
		return u.writeUpdatedConnections(ctx)
	}
	return nil
}

func (u *connectionStateTableUpdater) finishedUpdating(ctx context.Context) error {
	log.Printf("[WARN] finishedUpdating")
	return u.writeUpdatedConnections(ctx)
}

func (u *connectionStateTableUpdater) onConnectionDeleted(ctx context.Context, name string) error {
	u.deletedConnections = append(u.deletedConnections, name)
	if len(u.deletedConnections) >= updatePageSize {
		log.Printf("[WARN] onConnectionDeleted deleting page")
		return u.removeDeletedConnections(ctx)
	}
	return nil

}

func (u *connectionStateTableUpdater) finishedDeleting(ctx context.Context) error {
	log.Printf("[WARN] finishedUpdfinishedDeletingating")
	return u.removeDeletedConnections(ctx)
}

func (u *connectionStateTableUpdater) onConnectionError(ctx context.Context, connectionName string, err error) error {
	sql := getConnectionStateErrorSql(connectionName, err)
	if _, err := executeSqlAsRoot(ctx, sql); err != nil {
		return err
	}
	return nil
	// TODO KAI send notification
}

func (u *connectionStateTableUpdater) writeUpdatedConnections(ctx context.Context) error {
	// TODO KAI send notification
	var queries []string
	for _, c := range u.updatedConnections {
		queries = append(queries, getUpdateConnectionStateSql(c, constants.ConnectionStateReady))
	}
	if _, err := executeSqlAsRoot(ctx, queries...); err != nil {
		return err
	}
	// clear page of updated connections
	u.updatedConnections = nil
	return nil
}

func (u *connectionStateTableUpdater) removeDeletedConnections(ctx context.Context) error {
	var queries []string
	for _, c := range u.deletedConnections {
		queries = append(queries, getDeleteConnectionStateSql(c))
	}
	if _, err := executeSqlAsRoot(ctx, queries...); err != nil {
		return err
	}
	return nil
}

func getDeleteConnectionStateSql(connectionName string) string {
	return fmt.Sprintf(`delete from %s.%s where name = '%s'`,
		constants.InternalSchema, constants.ConnectionStateTable,
		connectionName)
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
