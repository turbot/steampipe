package steampipeconfig

import (
	"context"
	"fmt"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
)

// LoadConnectionState populates a ConnectionStateMap from the connection_state table
// it verifies the table has been initialised by calling RefreshConnections after db startup
func LoadConnectionState(ctx context.Context, conn *pgx.Conn, opts ...LoadConnectionStateOption) (ConnectionStateMap, error) {
	config := &LoadConnectionStateConfiguration{}
	for _, opt := range opts {
		opt(config)
	}

	// max duration depends on if waiting for ready or just pending
	// default value is if we are waiting for pending
	// set this to a long enough time for ConnectionUpdates to be generated for a large connection count
	// TODO this time can be reduced once all; plugins are using v5.4.1 of the sdk
	maxDuration := 1 * time.Minute
	retryInterval := 50 * time.Millisecond
	if config.WaitMode == WaitForReady || config.WaitMode == WaitForSearchPath {
		// is we are waiting for all connections to be ready, wait up to 10 minutes
		maxDuration = 10 * time.Minute
		retryInterval = 250 * time.Millisecond
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
				return retry.RetryableError(fmt.Errorf("connection state is pending"))
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
			// verify that no schemas are in error state
			// (this returns an error if any schemas are in error state)
			return checkConnectionErrors(config.Connections, connectionStateMap)

		}
		return nil

	})

	return connectionStateMap, err
}

func checkConnectionsAreReady(ctx context.Context, connectionStateMap ConnectionStateMap, config *LoadConnectionStateConfiguration) error {
	if !connectionStateMap.Loaded(config.Connections...) {
		statusMessage := GetLoadingConnectionStatusMessage(connectionStateMap, config.Connections...)
		statushooks.SetStatus(ctx, statusMessage)
		return retry.RetryableError(fmt.Errorf("connection state is still loading"))
	}
	return nil
}

// if any of the given connections are in error state, return an error
func checkConnectionErrors(schemas []string, connectionStateMap ConnectionStateMap) error {
	var errors []error
	for _, connectionName := range schemas {
		connectionState, ok := connectionStateMap[connectionName]
		if !ok {
			// not expected but not impossible - state may have changed while we iterate
			continue
		}
		if connectionState.State == constants.ConnectionStateError {
			err := fmt.Errorf("connection '%s' failed to load: %s",
				connectionName, typehelpers.SafeString(connectionState.ConnectionError))
			errors = append(errors, err)
		}
	}
	return error_helpers.CombineErrors(errors...)
}

func GetLoadingConnectionStatusMessage(connectionStateMap ConnectionStateMap, requiredSchemas ...string) string {
	var connectionSummary = connectionStateMap.GetSummary()

	readyCount := connectionSummary[constants.ConnectionStateReady]
	totalCount := len(connectionStateMap) - connectionSummary[constants.ConnectionStateDeleting]

	loadedMessage := fmt.Sprintf("Loaded %d of %d %s",
		readyCount,
		totalCount,
		utils.Pluralize("connection", totalCount))

	if len(requiredSchemas) == 0 {
		return loadedMessage
	}
	// TODO kai think about display of arrays
	return fmt.Sprintf("Waiting for %s '%s' to load (%s)", utils.Pluralize("connection", len(requiredSchemas)), strings.Join(requiredSchemas, "','"), loadedMessage)
}

func loadConnectionState(ctx context.Context, conn *pgx.Conn) (ConnectionStateMap, error) {
	query := fmt.Sprintf(`SELECT name,
		state,
		error,	
		plugin,
		schema_mode,
		schema_hash,
		connection_mod_time,
		plugin_mod_time
	FROM  %s.%s `, constants.InternalSchema, constants.ConnectionStateTable)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res = make(ConnectionStateMap)

	connectionDataList, err := pgx.CollectRows(rows, pgx.RowToStructByName[ConnectionState])
	if err != nil {
		return nil, err
	}

	for _, c := range connectionDataList {
		// copy into loop var
		connectionData := c
		// TODO remove this usage of GlobalConfig.Connections
		// (possibly remove connectionData.Connection altogether?)
		//https://github.com/turbot/steampipe/issues/3387
		// get connection config for this connection
		// (this will not be there for a deletion)
		connection, _ := GlobalConfig.Connections[connectionData.ConnectionName]

		connectionData.Connection = connection
		res[c.ConnectionName] = &connectionData
	}

	return res, nil
}

func SaveConnectionStateFile(res *RefreshConnectionResult, connectionUpdates *ConnectionUpdates) {
	// now serialise the connection state
	connectionState := make(ConnectionStateMap, len(connectionUpdates.FinalConnectionState))
	for k, v := range connectionUpdates.FinalConnectionState {
		connectionState[k] = v
	}
	// NOTE: add any connection which failed
	for c, reason := range res.FailedConnections {
		connectionState[c].State = constants.ConnectionStateError
		connectionState[c].SetError(reason)
	}
	for pluginName, connections := range connectionUpdates.MissingPlugins {
		// add in missing connections
		for _, c := range connections {
			connectionData := NewConnectionData(pluginName, &c, time.Now())
			connectionData.State = constants.ConnectionStateError
			connectionData.SetError(constants.ConnectionErrorPluginNotInstalled)
			connectionState[c.Name] = connectionData
		}
	}

	// update connection state and write the missing and failed plugin connections
	if err := connectionState.Save(); err != nil {
		res.Error = err
	}
}

func DeleteConnectionStateFile() {
	os.Remove(filepaths.ConnectionStatePath())
}
