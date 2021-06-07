package steampipeconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// GetConnectionState :: load connection state file, and remove any connections which do not exist in the db
func GetConnectionState(schemas []string) (ConnectionMap, error) {
	utils.LogTime("steampipeconfig.GetConnectionState start")
	defer utils.LogTime("steampipeconfig.GetConnectionState end")
	// load the connection state file and filter out any connections which are not in the list of schemas
	connectionState, err := loadConnectionStateFile()
	if err != nil {
		return nil, err
	}
	return pruneConnectionState(connectionState, schemas), nil
}

// load and parse the connection config
func loadConnectionStateFile() (ConnectionMap, error) {
	var connectionState ConnectionMap
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
		return nil, fmt.Errorf("error parsing %s: %v", connectionStatePath, err)
	}

	return connectionState, nil
}

// update connection map to remove any connections which do not exist in the list of given schemas
func pruneConnectionState(connections ConnectionMap, schemas []string) ConnectionMap {
	var actualConnectionState = make(ConnectionMap)
	for _, connectionName := range schemas {
		if connection, ok := connections[connectionName]; ok {
			actualConnectionState[connectionName] = connection
		}
	}

	return actualConnectionState
}

func SaveConnectionState(state ConnectionMap) error {
	return writeJson(state, constants.ConnectionStatePath())
}

func writeJson(data interface{}, path string) error {
	j, _ := json.MarshalIndent(data, "", " ")
	return ioutil.WriteFile(path, j, 0644)
}
