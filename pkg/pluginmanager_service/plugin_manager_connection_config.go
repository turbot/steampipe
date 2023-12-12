package pluginmanager_service

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe-plugin-sdk/v5/error_helpers"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"log"
)

func (m *PluginManager) getConnectionConfig(connectionName string) (*sdkproto.ConnectionConfig, error) {
	connectionConfig, ok := m.connectionConfigMap[connectionName]
	if !ok {
		return nil, fmt.Errorf("connection '%s' does not exist in connection config", connectionName)
	}
	return connectionConfig, nil
}

func (m *PluginManager) handleConnectionConfigChanges(ctx context.Context, newConfigMap map[string]*sdkproto.ConnectionConfig) error {
	// now determine whether there are any new or deleted connections
	addedConnections, deletedConnections, changedConnections := m.connectionConfigMap.Diff(newConfigMap)

	// build a map of UpdateConnectionConfig requests, keyed by plugin instance
	requestMap := make(map[string]*sdkproto.UpdateConnectionConfigsRequest)

	// for deleted connections, remove from plugins and pluginConnectionConfigs
	m.handleDeletedConnections(deletedConnections, requestMap)

	// for new connections, add to plugins and pluginConnectionConfigs
	m.handleAddedConnections(addedConnections, requestMap)
	// for updated connections just add to request map
	m.handleUpdatedConnections(changedConnections, requestMap)
	// update connectionConfigMap
	m.connectionConfigMap = newConfigMap

	// rebuild pluginConnectionConfigMap
	m.populatePluginConnectionConfigs()

	// now send UpdateConnectionConfigs for all update plugins
	return m.sendUpdateConnectionConfigs(requestMap)
}

func (m *PluginManager) sendUpdateConnectionConfigs(requestMap map[string]*sdkproto.UpdateConnectionConfigsRequest) error {
	var errors []error
	for pluginInstance, req := range requestMap {
		runningPlugin, pluginAlreadyRunning := m.runningPluginMap[pluginInstance]

		// if the pluginInstance is not running (or is not multi connection, so is not in this map), return
		if !pluginAlreadyRunning {
			continue
		}

		pluginClient, err := sdkgrpc.NewPluginClient(runningPlugin.client, runningPlugin.imageRef)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		err = pluginClient.UpdateConnectionConfigs(req)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return error_helpers.CombineErrors(errors...)
}

// this mutates requestMap
func (m *PluginManager) handleAddedConnections(addedConnections map[string][]*sdkproto.ConnectionConfig, requestMap map[string]*sdkproto.UpdateConnectionConfigsRequest) {
	// for new connections, add to plugins , pluginConnectionConfigs and connectionConfig
	// (but only if the plugin is already started - if not we do nothing here - refreshConnections will start the plugin)
	for p, connections := range addedConnections {
		// find the existing running plugin for this plugin
		// if this plugins is NOT running (or is not multi connection), skip here - we will start it when running refreshConnections
		runningPlugin, pluginAlreadyRunning := m.runningPluginMap[p]
		if !pluginAlreadyRunning {
			log.Printf("[TRACE] handleAddedConnections - plugin '%s' has been added to connection config and is not running - doing nothing here as it will be started by refreshConnections", p)
			continue
		}

		// get or create req for this plugin
		req, ok := requestMap[p]
		if !ok {
			req = &sdkproto.UpdateConnectionConfigsRequest{}
		}

		for _, connection := range connections {
			// add this connection to the running plugin
			runningPlugin.reattach.AddConnection(connection.Connection)

			// add to updateConnectionConfigsRequest
			req.Added = append(req.Added, connection)
		}
		// write back to map
		requestMap[p] = req
	}
}

// this mutates requestMap
func (m *PluginManager) handleDeletedConnections(deletedConnections map[string][]*sdkproto.ConnectionConfig, requestMap map[string]*sdkproto.UpdateConnectionConfigsRequest) {
	for p, connections := range deletedConnections {
		runningPlugin, pluginAlreadyRunning := m.runningPluginMap[p]
		if !pluginAlreadyRunning {
			continue
		}

		// get or create req for this plugin
		req, ok := requestMap[p]
		if !ok {
			req = &sdkproto.UpdateConnectionConfigsRequest{}
		}

		for _, connection := range connections {
			// remove this connection from the running plugin
			runningPlugin.reattach.RemoveConnection(connection.Connection)

			// add to updateConnectionConfigsRequest
			req.Deleted = append(req.Deleted, connection)
		}
		// write back to map
		requestMap[p] = req
	}
}

// this mutates requestMap
func (m *PluginManager) handleUpdatedConnections(updatedConnections map[string][]*sdkproto.ConnectionConfig, requestMap map[string]*sdkproto.UpdateConnectionConfigsRequest) {
	// for new connections, add to plugins , pluginConnectionConfigs and connectionConfig
	// (but only if the plugin is already started - if not we do nothing here - refreshConnections will start the plugin)
	for p, connections := range updatedConnections {
		// get or create req for this plugin
		req, ok := requestMap[p]
		if !ok {
			req = &sdkproto.UpdateConnectionConfigsRequest{}
		}

		// add to updateConnectionConfigsRequest
		req.Changed = append(req.Changed, connections...)
		// write back to map
		requestMap[p] = req
	}
}
