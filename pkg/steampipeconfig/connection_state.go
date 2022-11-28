package steampipeconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
)

// GetConnectionState loads the connection state file, and remove any connections which do not exist in the db
func GetConnectionState(schemaNames []string) (state ConnectionDataMap, stateModified bool, err error) {
	utils.LogTime("steampipeconfig.GetConnectionState start")
	defer utils.LogTime("steampipeconfig.GetConnectionState end")

	// load the connection state file and filter out any connections which are not in the list of schemas
	connectionState, err := loadConnectionStateFile()
	if err != nil {
		return nil, false, err
	}
	return pruneConnectionState(connectionState, schemaNames)
}

// load and parse the connection config
func loadConnectionStateFile() (ConnectionDataMap, error) {
	var connectionState ConnectionDataMap
	connectionStatePath := filepaths.ConnectionStatePath()

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

// update connection map to remove any connections which do not exist in the list of given schemas
// if this function removes connections, set stateModified to true
func pruneConnectionState(connectionState ConnectionDataMap, schemaNames []string) (prunedState ConnectionDataMap, stateModified bool, err error) {
	prunedState = make(ConnectionDataMap)
	for _, connectionName := range schemaNames {
		if connection, ok := connectionState[connectionName]; ok {
			prunedState[connectionName] = connection
		}

	}
	stateModified = len(connectionState) != len(prunedState)
	return
}
