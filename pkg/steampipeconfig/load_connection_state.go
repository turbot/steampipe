package steampipeconfig

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
)

// LoadConnectionState populates a ConnectionStateMap from the connection_state table
// it verifies the table has been initialised by calling RefreshConnections after db startup
func LoadConnectionState(ctx context.Context, conn *pgx.Conn, opts ...LoadConnectionStateOption) (ConnectionStateMap, error) {
	log.Println("[DEBUG] LoadConnectionState start")
	defer log.Println("[DEBUG] LoadConnectionState end")

	config := &LoadConnectionStateConfiguration{}
	for _, opt := range opts {
		opt(config)
	}

	// max duration depends on if waiting for ready or just pending
	// default value is if we are waiting for pending
	// set this to a long enough time for ConnectionUpdates to be generated for a large connection count
	// TODO this time can be reduced once all; plugins are using v5.4.1 of the sdk
	maxDuration := 1 * time.Minute
	retryInterval := 250 * time.Millisecond
	if config.WaitMode == WaitForReady || config.WaitMode == WaitForSearchPath {
		// is we are waiting for all connections to be ready, wait up to 10 minutes
		maxDuration = 10 * time.Minute
	}
	backoff := retry.NewConstant(retryInterval)

	var connectionStateMap ConnectionStateMap

	err := retry.Do(ctx, retry.WithMaxDuration(maxDuration, backoff), func(ctx context.Context) error {
		var loadErr error
		connectionStateMap, loadErr = loadConnectionState(ctx, conn)
		if loadErr != nil {
			return loadErr
		}

		// now process any load options
		switch config.WaitMode {
		case WaitForReady:
			return checkConnectionsAreReady(ctx, connectionStateMap, config)
		case WaitForLoading:
			if connectionStateMap.Pending() {
				return retry.RetryableError(fmt.Errorf("timed out waiting for connection state to be updated from pending"))
			}
		case WaitForSearchPath:
			if len(config.SearchPath) == 0 {
				// nothing to do
				return nil
			}

			// wait for search path is called with a search path set - we must convert this into a set of
			// connections which we must wait for (the first connection for each plugin)
			// the first time we load the connection state, determine the connections we need to wait for
			if len(config.Connections) == 0 {
				// build list of connections we must wait for as update config
				config.Connections = connectionStateMap.GetFirstSearchPathConnectionForPlugins(config.SearchPath)
			}
			// now check if these connections are ready
			if err := checkConnectionsAreReady(ctx, connectionStateMap, config); err != nil {
				return err
			}

			// so all required connections are loaded, either 'ready' or 'error'
			// verify that not all schemas are in error state
			// (this returns an error if any schemas are in error state)
			if allConnectionsInError(config.Connections, connectionStateMap) {
				return fmt.Errorf("all connections in search path are in error")
			}
			return nil

		}
		return nil

	})

	return connectionStateMap, err
}

func loadConnectionState(ctx context.Context, conn *pgx.Conn, opts ...loadConnectionStateOption) (ConnectionStateMap, error) {
	config := &loadConnectionStateConfig{}
	for _, configOption := range opts {
		configOption(config)
	}
	log.Println("[TRACE] with config", config)

	var res = make(ConnectionStateMap)

	query := fmt.Sprintf(
		`select * FROM %s.%s `,
		constants.InternalSchema,
		constants.ConnectionTable,
	)
	legacyQuery := fmt.Sprintf(
		`select * FROM %s.%s `,
		constants.InternalSchema,
		constants.LegacyConnectionStateTable,
	)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		if !db_common.IsRelationNotFoundError(err) {
			return nil, err
		}
		// so it was a relation not found - try with legacy table
		rows, err = conn.Query(ctx, legacyQuery)
		if err != nil {
			return nil, err
		}
	}

	defer rows.Close()

	connectionStateList, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[ConnectionState])
	if err != nil {
		return nil, err
	}

	// convert to pointer arrau
	for _, c := range connectionStateList {
		// copy into loop var
		connectionState := c
		res[c.ConnectionName] = &connectionState
	}

	return res, nil
}

func checkConnectionsAreReady(ctx context.Context, connectionStateMap ConnectionStateMap, config *LoadConnectionStateConfiguration) error {
	if !connectionStateMap.Loaded(config.Connections...) {
		statusMessage := GetLoadingConnectionStatusMessage(connectionStateMap, config.Connections...)
		statushooks.SetStatus(ctx, statusMessage)
		return retry.RetryableError(fmt.Errorf("connection state is still loading"))
	}
	return nil
}

func allConnectionsInError(connectionsNames []string, connectionStateMap ConnectionStateMap) bool {
	if len(connectionsNames) == 0 {
		return false
	}
	for _, connectionName := range connectionsNames {
		connectionState, ok := connectionStateMap[connectionName]
		if !ok {
			// not expected but not impossible - state may have changed while we iterate
			continue
		}
		if connectionState.State != constants.ConnectionStateError {
			return false
		}
	}

	return true
}

func GetLoadingConnectionStatusMessage(connectionStateMap ConnectionStateMap, requiredSchemas ...string) string {
	var connectionSummary = connectionStateMap.GetSummary()

	readyCount := connectionSummary[constants.ConnectionStateReady]
	totalCount := len(connectionStateMap) - connectionSummary[constants.ConnectionStateDeleting]

	loadedMessage := fmt.Sprintf("Loaded %d of %d %s",
		readyCount,
		totalCount,
		utils.Pluralize("connection", totalCount))

	if len(requiredSchemas) == 1 {
		// if we are only waiting for a single schema, include that in the message
		return fmt.Sprintf("Waiting for connection '%s' to load (%s)", requiredSchemas[0], loadedMessage)
	}

	return loadedMessage
}

func SaveConnectionStateFile(res *RefreshConnectionResult, connectionUpdates *ConnectionUpdates) {
	// now serialise the connection state
	connectionState := make(ConnectionStateMap, len(connectionUpdates.FinalConnectionState))
	for k, v := range connectionUpdates.FinalConnectionState {
		connectionState[k] = v
	}
	// NOTE: add any connection which failed
	for c, reason := range res.FailedConnections {
		connectionState[c].SetError(reason)
	}

	// update connection state and write the missing and failed plugin connections
	if err := connectionState.Save(); err != nil {
		res.Error = err
	}
}

func DeleteConnectionStateFile() {
	os.Remove(filepaths.ConnectionStatePath())
}

type loadConnectionStateConfig struct {
}

type loadConnectionStateOption func(l *loadConnectionStateConfig)
