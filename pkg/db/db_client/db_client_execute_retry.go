package db_client

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sethvargo/go-retry"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

// execute query - if it fails with a "relation not found" error, determine whether this is because the required schema
// has not yet loaded and if so, wait for it to load and retry
func (c *DbClient) startQueryWithRetries(ctx context.Context, session *db_common.DatabaseSession, query string, args ...any) (pgx.Rows, error) {
	// long timeout to give refresh connections a chance to finish
	maxDuration := 10 * time.Minute
	backoffInterval := 250 * time.Millisecond
	backoff := retry.NewConstant(backoffInterval)

	conn := session.Connection.Conn()

	var res pgx.Rows
	err := retry.Do(ctx, retry.WithMaxDuration(maxDuration, backoff), func(ctx context.Context) error {
		rows, queryError := c.startQuery(ctx, conn, query, args...)
		// if there is no error, just return
		if queryError == nil {
			statushooks.SetStatus(ctx, "Loading results…")
			res = rows
			return nil
		}

		log.Printf("[WARN] startQueryWithRetries query error: %s", queryError.Error())

		// so there is an error - is it "relation not found"?
		missingSchema, _, relationNotFound := IsRelationNotFoundError(queryError)
		if !relationNotFound {
			log.Printf("[WARN] NOT relationNotFound - just returnin error")

			// just return it
			return queryError
		}
		// so this _was_ a "relation not found" error
		// load the connection state and connection config to see if the missing schema is in there at all
		// if there was a schema not found with an unqualified query, we keep trying until
		// the first search path schema for each plugin has loaded

		log.Printf("[WARN] relationNotFound - loading connection state")
		connectionStateMap, stateErr := steampipeconfig.LoadConnectionState(ctx, conn, steampipeconfig.WithWaitUntilLoading())
		if stateErr != nil {
			log.Printf("[WARN] >> stateErr: queryError - %s", stateErr.Error())
			// just return the query error
			return queryError
		}

		// if there are no connections, just return the error
		if len(connectionStateMap) == 0 {
			log.Printf("[WARN] connection state empty - returnin gquery error")
			return queryError
		}

		// is this an unqualified query...
		if missingSchema == "" {
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
			if _, err := steampipeconfig.LoadConnectionState(ctx, conn, steampipeconfig.WithWaitForSearchPath(searchPath)); err != nil {
				return err
			}

			// so now the connections are loaded - retry the query
			return retry.RetryableError(queryError)
		}

		// so a schema was specified
		// verify it exists in the connection state and is not disabled
		connectionState, missingSchemaExistsInStateMap := connectionStateMap[missingSchema]
		if !missingSchemaExistsInStateMap || connectionState.Disabled() {
			log.Printf("[WARN] missing schema is not in connection state map - just return the error. missingSchemaExistsInStateMap: %v, connectionState.Disabled() %v", missingSchema, connectionState.Disabled())
			//, missing schema is not in connection state map - just return the error
			return queryError
		}

		// so schema _is_ in the state map

		// if the connection is ready (and has been for more than the backoff interval) , just return the relation not found error
		if connectionState.State == constants.ConnectionStateReady && time.Since(connectionState.ConnectionModTime) > backoffInterval {
			log.Printf("[WARN] connection exists in ready state - just returning query error")
			return queryError
		}

		// if connection is in error,return the connection error
		if connectionState.State == constants.ConnectionStateError {
			log.Printf("[WARN] connection exists in error state - returning connection error")
			return fmt.Errorf("connection %s failed to load: %s", missingSchema, typehelpers.SafeString(connectionState.ConnectionError))
		}

		log.Printf("[WARN] connection exists but is not ready - waiting")

		// ok so we will retry
		// build the status message to display with a spinner, if needed
		statusMessage := steampipeconfig.GetLoadingConnectionStatusMessage(connectionStateMap, missingSchema)
		statushooks.SetStatus(ctx, statusMessage)
		return retry.RetryableError(queryError)
	})

	return res, err
}

func IsRelationNotFoundError(err error) (string, string, bool) {
	if err == nil {
		return "", "", false
	}
	pgErr, ok := err.(*pgconn.PgError)
	if !ok || pgErr.Code != "42P01" {
		return "", "", false
	}

	r := regexp.MustCompile(`^relation "(.*)\.(.*)" does not exist$`)
	captureGroups := r.FindStringSubmatch(pgErr.Message)
	if len(captureGroups) == 3 {

		return captureGroups[1], captureGroups[2], true
	}

	// maybe there is no schema
	r = regexp.MustCompile(`^relation "(.*)" does not exist$`)
	captureGroups = r.FindStringSubmatch(pgErr.Message)
	if len(captureGroups) == 2 {
		return "", captureGroups[1], true
	}
	return "", "", true
}
