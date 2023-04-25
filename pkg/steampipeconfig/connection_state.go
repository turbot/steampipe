package steampipeconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/sethvargo/go-retry"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
	"os"
	"time"
)

type LoadConnectionStateConfiguration struct {
	WaitForPending bool
}

type LoadConnectionStateOption = func(config *LoadConnectionStateConfiguration)

var WithWaitForPending = func(config *LoadConnectionStateConfiguration) {
	config.WaitForPending = true
}

// LoadConnectionState populates a ConnectionDataMap from the connection_state table
// it verifies the table has been initialised by calling RefreshConnections after db startup
func LoadConnectionState(ctx context.Context, conn *pgx.Conn, opts ...LoadConnectionStateOption) (ConnectionDataMap, error) {
	config := &LoadConnectionStateConfiguration{}
	for _, opt := range opts {
		opt(config)
	}
	if !config.WaitForPending {
		return loadConnectionState(ctx, conn)
	}

	maxAttempts := uint64(10)
	retryInterval := 50 * time.Millisecond
	backoff := retry.NewConstant(retryInterval * time.Millisecond)

	var connectionState ConnectionDataMap

	err := retry.Do(ctx, retry.WithMaxRetries(maxAttempts, backoff), func(ctx context.Context) error {
		var loadErr error
		connectionState, loadErr = loadConnectionState(ctx, conn)
		if loadErr == nil && connectionState.Pending() {
			loadErr = retry.RetryableError(fmt.Errorf("connection state is pending"))
		}
		return loadErr
	})

	return connectionState, err
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

		connectionData.StructVersion = ConnectionDataStructVersion
		connectionData.Connection = connection
		res[c.ConnectionName] = &connectionData
	}

	return res, nil
}

// LoadConnectionStateFile loads the connection state file
func LoadConnectionStateFile() (state ConnectionDataMap, err error) {
	utils.LogTime("LoadConnectionStateFile start")
	defer utils.LogTime("LoadConnectionStateFile end")

	var connectionState ConnectionDataMap
	connectionStatePath := filepaths.ConnectionStatePath()

	// if file does not exist, return empty struct
	if !filehelpers.FileExists(connectionStatePath) {
		return connectionState, nil
	}
	jsonFile, err := os.ReadFile(connectionStatePath)
	if err != nil {
		return nil, fmt.Errorf("error loading %s: %v", connectionStatePath, err)
	}

	err = json.Unmarshal(jsonFile, &connectionState)
	if err != nil {
		log.Printf("[TRACE] error parsing %s: %v", connectionStatePath, err)
		// If we fail to parse the state file, suppress the error and return an empty state
		// This will force the connection to refresh
		return make(ConnectionDataMap), nil
	}

	// check whether the loaded state file has an older struct version
	// this indicates that we need to refresh this connection - so remove the connection data from the map
	// (typically this would be used if we need to force a refresh of connection config,
	// for example if there is an update to the Postgres schema building code)
	for key, connectionData := range connectionState {
		if connectionData.StructVersion < ConnectionDataStructVersion {
			delete(connectionState, key)
		}
	}
	return connectionState, nil
}

func SaveConnectionStateFile(res *RefreshConnectionResult, connectionUpdates *ConnectionUpdates) {
	// now serialise the connection state
	connectionState := make(ConnectionDataMap, len(connectionUpdates.FinalConnectionState))
	for k, v := range connectionUpdates.FinalConnectionState {
		connectionState[k] = v
	}
	// NOTE: add any connection which failed
	for c := range res.FailedConnections {
		connectionState[c].State = constants.ConnectionStateError
		connectionState[c].SetError(constants.ConnectionErrorPluginFailedToStart)
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
