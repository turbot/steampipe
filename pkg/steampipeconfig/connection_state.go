package steampipeconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
	"os"
)

func LoadConnectionState(ctx context.Context, pool *pgxpool.Pool) (state ConnectionDataMap, err error) {
	query := fmt.Sprintf(`SELECT name,
		state,
		error,	
		plugin,
		schema_mode,
		schema_hash,
		plugin_mod_time
	FROM  %s.%s `, constants.InternalSchema, constants.ConnectionStateTable)

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var res = make(ConnectionDataMap)

	connectionDataList, err := pgx.CollectRows(rows, pgx.RowToStructByName[ConnectionData])
	if err != nil {
		return nil, err
	}

	for _, c := range connectionDataList {
		// get connection config for this connection
		connection, _ := GlobalConfig.Connections[c.ConnectionName]
		// this will not be there for a deletion

		c.StructVersion = ConnectionDataStructVersion
		c.Connection = connection
		res[c.ConnectionName] = &c
	}

	return res, nil
}

// LoadConnectionStateFile loads the connection state file
func LoadConnectionStateFile() (state ConnectionDataMap, err error) {
	utils.LogTime("steampipeconfig.LoadConnectionStateFile start")
	defer utils.LogTime("steampipeconfig.LoadConnectionStateFile end")

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
