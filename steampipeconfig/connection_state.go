package steampipeconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
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
	connectionStatePath := constants.ConnectionStatePath()

	if !helpers.FileExists(connectionStatePath) {
		return connectionState, nil
	}
	jsonFile, err := ioutil.ReadFile(connectionStatePath)
	if err != nil {
		return nil, fmt.Errorf("error loading %s: %v", connectionStatePath, err)
	}

	err = json.Unmarshal(jsonFile, &connectionState)
	if err != nil {
		log.Printf("[TRACE] error parsing %s: %v", connectionStatePath, err)
		// If we fail to parse the state file, suppress the error and return an empty state
		// This will force the connection to refresh
		return make (ConnectionDataMap), nil
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
	return writeJson(state, constants.ConnectionStatePath())
}

func writeJson(data interface{}, path string) error {
	j, _ := json.MarshalIndent(data, "", " ")
	return ioutil.WriteFile(path, j, 0644)
}
