package db_client

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sethvargo/go-retry"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/connection_sync"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"regexp"
	"time"
)

// execute query - if it fails with a "relation not found" error, determine whether this is because the required schema
// has not yet loaded and if so, wait for it to load and retry
func (c *DbClient) startQueryWithRetries(ctx context.Context, conn *pgx.Conn, query string, args ...any) (pgx.Rows, error) {
	maxDuration := 10 * time.Minute
	backoffInterval := 250 * time.Millisecond
	backoff := retry.NewConstant(backoffInterval)

	var res pgx.Rows
	err := retry.Do(ctx, retry.WithMaxDuration(maxDuration, backoff), func(ctx context.Context) error {
		rows, queryError := c.startQuery(ctx, conn, query, args...)
		if queryError == nil {
			statushooks.SetStatus(ctx, "Loading results...")
			res = rows
			return nil
		}

		missingSchema, _, relationNotFound := isRelationNotFoundError(queryError)
		if !relationNotFound {
			return queryError
		}
		// so this _was_ a relation not found error
		// load the connection state and connection config to see if the missing schema is in there at all
		// if there was a schema not found with an unqualified query, we keep trying until ALL the schemas have loaded

		connectionStateMap, stateErr := steampipeconfig.LoadConnectionState(ctx, conn, steampipeconfig.WithWaitForPending())
		if stateErr != nil {
			// just return the query error
			return queryError
		}
		// if there are no connections, just return the error
		if len(connectionStateMap) == 0 {
			return queryError
		}

		if missingSchema == "" {
			// if all connections are ready (and have been for more than the backoff interval) , just return the relation not found error
			if connectionStateMap.Loaded() && time.Since(connectionStateMap.ConnectionModTime()) > backoffInterval {
				return queryError
			}

			// TODO KAI test this
			// otherwise we need to wait for the first schema of everything plugin to load
			if err := connection_sync.WaitForSearchPathHeadSchemas(ctx, c, c.GetRequiredSessionSearchPath()); err != nil {
				return err
			}
			return nil
		}

		// so a schema was specified - verify it exists in the connection state
		connectionState, missingSchemaExistsInStateMap := connectionStateMap[missingSchema]
		if missingSchemaExistsInStateMap {
			// if the connection is ready (and has been for more than the backoff interval) , just return the relation not found error
			if connectionState.State == constants.ConnectionStateReady && time.Since(connectionState.ConnectionModTime) > backoffInterval {
				return queryError
			}
			// if connection is in error and there is connection error
			if connectionState.State == constants.ConnectionStateError {
				return fmt.Errorf("connection %s failed to load: %s", missingSchema, typehelpers.SafeString(connectionState.ConnectionError))
			}
			// retry
			// build the status message to display with a spinner, if needed
			statusMessage := steampipeconfig.GetLoadingConnectionStatusMessage(connectionStateMap, missingSchema)
			statushooks.SetStatus(ctx, statusMessage)
			return retry.RetryableError(queryError)
		}

		// otherwise, missing schema is not in connection state map - just return the error
		return queryError

	})

	return res, err
}

func isRelationNotFoundError(err error) (string, string, bool) {
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
