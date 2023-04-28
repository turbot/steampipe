package db_client

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
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

		connectionStateMap, stateErr := steampipeconfig.LoadConnectionState(ctx, conn, steampipeconfig.WithWaitForPending)
		if stateErr != nil {
			// just return the query error
			return queryError
		}
		// if there are no connections, just return the error
		if len(connectionStateMap) == 0 {
			return queryError
		}

		statusMessage := getLoadingConnectionStatusMessage(connectionStateMap, missingSchema)

		// if a schema was specified, verify it exists in the connection state or connection config

		if missingSchema != "" {
			connectionState, missingSchemaExistsInStateMap := connectionStateMap[missingSchema]
			if missingSchemaExistsInStateMap {
				// if connection is in error or has been ready for more than the backoff interval, do not retry
				// (in other words, if it has only just become ready, then retry the query)
				if connectionState.State == constants.ConnectionStateError ||
					connectionState.State == constants.ConnectionStateReady && time.Since(connectionState.ConnectionModTime) > backoffInterval {
					return queryError
				}
			} else {
				// missing schema is not in connection state map - just return the error
				return queryError
			}
		} else {
			// if no schema was specified, return if the connection state is not pending
			if !connectionStateMap.Pending() {
				return queryError
			}

			// otherwise we need to wait for everything to load
		}

		statushooks.SetStatus(ctx, statusMessage)

		// retry
		return retry.RetryableError(queryError)
	})

	return res, err
}

func getLoadingConnectionStatusMessage(connectionStateMap steampipeconfig.ConnectionDataMap, missingSchema string) string {
	var connectionSummary = connectionStateMap.GetSummary()

	readyCount := connectionSummary[constants.ConnectionStateReady]
	totalCount := len(connectionStateMap) - connectionSummary[constants.ConnectionStateDeleting]

	loadedMessage := fmt.Sprintf("Loaded %d of %d %s",
		readyCount,
		totalCount,
		utils.Pluralize("connection", totalCount))

	if missingSchema == "" {
		return loadedMessage
	}

	return fmt.Sprintf("Waiting for connection '%s' to load (%s)", missingSchema, loadedMessage)
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
