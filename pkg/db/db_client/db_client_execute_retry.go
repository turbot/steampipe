package db_client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sethvargo/go-retry"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// execute query - if it fails with a "relation not found" error, determine whether this is because the required schema
// has not yet loaded and if so, wait for it to load and retry
func (c *DbClient) startQueryWithRetries(ctx context.Context, session *db_common.DatabaseSession, query string, args ...any) (pgx.Rows, error) {
	log.Println("[TRACE] DbClient.startQueryWithRetries start")
	defer log.Println("[TRACE] DbClient.startQueryWithRetries end")

	// long timeout to give refresh connections a chance to finish
	maxDuration := 10 * time.Minute
	backoffInterval := 250 * time.Millisecond
	backoff := retry.NewConstant(backoffInterval)

	conn := session.Connection.Conn()

	var res pgx.Rows
	count := 0
	err := retry.Do(ctx, retry.WithMaxDuration(maxDuration, backoff), func(ctx context.Context) error {
		count++
		log.Println("[TRACE] starting", count)
		rows, queryError := c.startQuery(ctx, conn, query, args...)
		// if there is no error, just return
		if queryError == nil {
			log.Println("[TRACE] no queryError")
			statushooks.SetStatus(ctx, "Loading results…")
			res = rows
			return nil
		}

		log.Println("[TRACE] queryError:", queryError)
		// so there is an error - is it "relation not found"?
		missingSchema, _, relationNotFound := db_common.GetMissingSchemaFromIsRelationNotFoundError(queryError)
		if !relationNotFound {
			log.Println("[TRACE] queryError not relation not found")
			// just return it
			return queryError
		}

		// get a connection from the system pool to query the connection state table
		sysConn, err := c.managementPool.Acquire(ctx)
		if err != nil {
			return retry.RetryableError(err)
		}
		defer sysConn.Release()
		// so this _was_ a "relation not found" error
		// load the connection state and connection config to see if the missing schema is in there at all
		// if there was a schema not found with an unqualified query, we keep trying until
		// the first search path schema for each plugin has loaded
		connectionStateMap, stateErr := steampipeconfig.LoadConnectionState(ctx, sysConn.Conn(), steampipeconfig.WithWaitUntilLoading())
		if stateErr != nil {
			log.Println("[TRACE] could not load connection state map:", stateErr)
			// just return the query error
			return queryError
		}

		// if there are no connections, just return the error
		if len(connectionStateMap) == 0 {
			log.Println("[TRACE] no data in connection state map")
			return queryError
		}

		// is this an unqualified query...
		if missingSchema == "" {
			log.Println("[TRACE] this was an unqualified query")
			// refresh the search path, as now the connection state is in loading state, search paths may have been updated
			if err := c.ensureSessionSearchPath(ctx, session); err != nil {
				return queryError
			}

			// we need the first search path connection for each plugin to be loaded
			searchPath := c.GetRequiredSessionSearchPath()
			requiredConnections := connectionStateMap.GetFirstSearchPathConnectionForPlugins(searchPath)
			// if required connections are ready (and have been for more than the backoff interval) , just return the relation not found error
			if connectionStateMap.Loaded(requiredConnections...) && time.Since(connectionStateMap.ConnectionModTime()) > backoffInterval {
				return queryError
			}

			// otherwise we need to wait for the first schema of everything plugin to load
			if _, err := steampipeconfig.LoadConnectionState(ctx, sysConn.Conn(), steampipeconfig.WithWaitForSearchPath(searchPath)); err != nil {
				return err
			}

			// so now the connections are loaded - retry the query
			return retry.RetryableError(queryError)
		}

		// so a schema was specified
		// verify it exists in the connection state and is not disabled
		connectionState, missingSchemaExistsInStateMap := connectionStateMap[missingSchema]
		if !missingSchemaExistsInStateMap {
			log.Println("[TRACE] schema", missingSchema, "is not in schema map")
			//, missing schema is not in connection state map - just return the error
			return queryError
		}

		// so schema _is_ in the state map
		if connectionState.Disabled() {
			log.Println("[TRACE] schema", missingSchema, "is disabled")
			return queryError
		}

		// if the connection is ready (and has been for more than the backoff interval) , just return the relation not found error
		if connectionState.State == constants.ConnectionStateReady && time.Since(connectionState.ConnectionModTime) > backoffInterval {
			log.Println("[TRACE] schema", missingSchema, "has been ready for a long time")
			return queryError
		}

		// if connection is in error,return the connection error
		if connectionState.State == constants.ConnectionStateError {
			log.Println("[TRACE] schema", missingSchema, "is in error")
			return fmt.Errorf("connection %s failed to load: %s", missingSchema, typehelpers.SafeString(connectionState.ConnectionError))
		}

		// ok so we will retry
		// build the status message to display with a spinner, if needed
		statusMessage := steampipeconfig.GetLoadingConnectionStatusMessage(connectionStateMap, missingSchema)
		statushooks.SetStatus(ctx, statusMessage)
		return retry.RetryableError(queryError)
	})

	return res, err
}
