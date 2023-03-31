package db_local

import (
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"os"
	"time"
)

func deleteConnectionState() {
	os.Remove(filepaths.ConnectionStatePath())
}

func serialiseConnectionState(res *steampipeconfig.RefreshConnectionResult, connectionUpdates *steampipeconfig.ConnectionUpdates) {
	// now serialise the connection state
	connectionState := make(steampipeconfig.ConnectionDataMap, len(connectionUpdates.RequiredConnectionState))
	for k, v := range connectionUpdates.RequiredConnectionState {
		connectionState[k] = v
	}
	// NOTE: add any connection which failed
	for c := range res.FailedConnections {
		connectionState[c].ConnectionState = constants.ConnectionStateError
		connectionState[c].SetError(constants.ConnectionErrorPluginFailedToStart)
	}
	for pluginName, connections := range connectionUpdates.MissingPlugins {
		// add in missing connections
		for _, c := range connections {
			connectionData := steampipeconfig.NewConnectionData(pluginName, &c, time.Now())
			connectionData.ConnectionState = constants.ConnectionStateError
			connectionData.SetError(constants.ConnectionErrorPluginNotInstalled)
			connectionState[c.Name] = connectionData
		}
	}

	// update connection state and write the missing and failed plugin connections
	if err := connectionState.Save(); err != nil {
		res.Error = err
	}
}
