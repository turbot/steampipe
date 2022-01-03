package steampipeconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/file_paths"
	"github.com/turbot/steampipe/utils"
)

// GetConnectionState loads the connection state file, and remove any connections which do not exist in the db
func GetConnectionState(schemaNames []string) (ConnectionDataMap, error) {
	utils.LogTime("steampipeconfig.GetConnectionState start")
	defer utils.LogTime("steampipeconfig.GetConnectionState end")

	// load the connection state file and filter out any connections which are not in the list of schemas
	connectionState, err := loadConnectionStateFile()
	if err != nil {
		return nil, err
	}
	return pruneConnectionState(connectionState, schemaNames), nil
}

// load and parse the connection config
func loadConnectionStateFile() (ConnectionDataMap, error) {
	var connectionState ConnectionDataMap
	connectionStatePath := file_paths.ConnectionStatePath()

	if !helpers.FileExists(connectionStatePath) {
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

// update connection map to remove any connections which do not exist in the list of given schemas
func pruneConnectionState(connections ConnectionDataMap, schemaNames []string) ConnectionDataMap {
	var actualConnectionState = make(ConnectionDataMap)
	for _, connectionName := range schemaNames {
		if connection, ok := connections[connectionName]; ok {
			actualConnectionState[connectionName] = connection
		}
	}

	return actualConnectionState
}

func SaveConnectionState(state ConnectionDataMap) error {
	return writeJson(state, file_paths.ConnectionStatePath())
}

func writeJson(data interface{}, path string) error {
	j, _ := json.MarshalIndent(data, "", " ")
	return os.WriteFile(path, j, 0644)
}
