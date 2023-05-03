package steampipeconfig

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
)

//type ConnectionState ConnectionDataMap

type LoadConnectionStateConfiguration struct {
	WaitForPending bool
	WaitForReady   bool
	Connections    []string
}

type LoadConnectionStateOption = func(config *LoadConnectionStateConfiguration)

var WithWaitForPending = func() func(config *LoadConnectionStateConfiguration) {
	return func(config *LoadConnectionStateConfiguration) {
		config.WaitForPending = true
	}
}
var WithWaitUntilReady = func(connections ...string) func(config *LoadConnectionStateConfiguration) {
	return func(config *LoadConnectionStateConfiguration) {
		config.Connections = connections
		config.WaitForReady = true
	}
}

// LoadConnectionState populates a ConnectionDataMap from the connection_state table
// it verifies the table has been initialised by calling RefreshConnections after db startup
func LoadConnectionState(ctx context.Context, conn *pgx.Conn, opts ...LoadConnectionStateOption) (ConnectionDataMap, error) {
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
	if config.WaitForReady {
		// is we are waiting for all connections to be ready, wait up to 10 minutes
		maxDuration = 10 * time.Minute
		retryInterval = 250 * time.Millisecond
	}
	backoff := retry.NewConstant(retryInterval)

	var connectionState ConnectionDataMap

	err := retry.Do(ctx, retry.WithMaxDuration(maxDuration, backoff), func(ctx context.Context) error {
		var loadErr error
		connectionState, loadErr = loadConnectionState(ctx, conn)
		if loadErr == nil {
			if config.WaitForReady && !connectionState.Loaded(config.Connections...) {
				statusMessage := GetLoadingConnectionStatusMessage(connectionState, config.Connections...)
				statushooks.SetStatus(ctx, statusMessage)
				loadErr = retry.RetryableError(fmt.Errorf("connection state is still loading"))
			} else if config.WaitForPending && connectionState.Pending() {
				loadErr = retry.RetryableError(fmt.Errorf("connection state is pending"))
			}
		}
		return loadErr
	})

	return connectionState, err
}

func GetLoadingConnectionStatusMessage(connectionStateMap ConnectionDataMap, requiredSchemas ...string) string {
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

func loadConnectionState(ctx context.Context, conn *pgx.Conn) (ConnectionDataMap, error) {
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

	var res = make(ConnectionDataMap)

	connectionDataList, err := pgx.CollectRows(rows, pgx.RowToStructByName[ConnectionData])
	if err != nil {
		return nil, err
	}

	for _, c := range connectionDataList {
		// copy into loop var
		connectionData := c
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
	connectionState := make(ConnectionDataMap, len(connectionUpdates.FinalConnectionState))
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
