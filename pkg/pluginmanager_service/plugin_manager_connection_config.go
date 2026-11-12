package pluginmanager_service

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe-plugin-sdk/v6/error_helpers"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v6/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/connection"
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

// applyChangesMadeWhileStarting sends a plugin which has just started the changes made to its connections
// since startedWith, the connection configs it was started with, and the user rate limiters if they changed.
// Those changes were skipped while the plugin was starting.
// NOTE: caller must hold m.mut lock
func (m *PluginManager) applyChangesMadeWhileStarting(p *runningPlugin, startedWith []*sdkproto.ConnectionConfig) {
	requestMap := m.requestsForChangesMadeWhileStarting(p.pluginInstance, startedWith)
	if err := m.sendUpdateConnectionConfigs(requestMap); err != nil {
		log.Printf("[WARN] applyChangesMadeWhileStarting - failed to send connection config changes to plugin '%s': %s", p.pluginInstance, err.Error())
	}

	if p.limitersChangedWhileStarting && p.reattach.SupportedOperations.RateLimiters {
		pluginClient, err := sdkgrpc.NewPluginClient(p.client, p.imageRef)
		if err == nil {
			err = m.setRateLimitersInternal(p.pluginInstance, pluginClient)
		}
		if err != nil {
			log.Printf("[WARN] applyChangesMadeWhileStarting - failed to send rate limiters to plugin '%s': %s", p.pluginInstance, err.Error())
		}
	}
}

// requestsForChangesMadeWhileStarting builds the update requests which bring a plugin started with startedWith
// up to date with the current connection config
// NOTE: caller must hold m.mut lock
func (m *PluginManager) requestsForChangesMadeWhileStarting(pluginInstance string, startedWith []*sdkproto.ConnectionConfig) map[string]*sdkproto.UpdateConnectionConfigsRequest {
	startedWithMap := make(connection.ConnectionConfigMap, len(startedWith))
	for _, c := range startedWith {
		startedWithMap[c.Connection] = c
	}
	currentMap := make(connection.ConnectionConfigMap)
	for _, c := range m.pluginConnectionConfigMap[pluginInstance] {
		currentMap[c.Connection] = c
	}

	added, deleted, changed := startedWithMap.Diff(currentMap)
	requestMap := make(map[string]*sdkproto.UpdateConnectionConfigsRequest)
	m.handleDeletedConnections(deleted, requestMap)
	m.handleAddedConnections(added, requestMap)
	m.handleUpdatedConnections(changed, requestMap)
	return requestMap
}

func (m *PluginManager) sendUpdateConnectionConfigs(requestMap map[string]*sdkproto.UpdateConnectionConfigsRequest) error {
	var errors []error
	for pluginInstance, req := range requestMap {
		runningPlugin, pluginAlreadyRunning := m.runningPluginMap[pluginInstance]

		// if the pluginInstance is not running (or is not multi connection, so is not in this map), return
		if !pluginAlreadyRunning {
			continue
		}
		// if the plugin is still starting there is no client to send the update to
		// (the plugin was started with the connection configs read before this change -
		// startPlugin sends it the changes made since, once it has started)
		if runningPlugin.isStarting() {
			log.Printf("[DEBUG] sendUpdateConnectionConfigs - plugin '%s' is still starting - connection config update will be sent when it has started", pluginInstance)
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
		// if the plugin is still starting it has no reattach config yet - treat it as not running
		// (the plugin was started with the connection configs read before this change -
		// startPlugin sends it the changes made since, once it has started)
		if runningPlugin.isStarting() {
			log.Printf("[DEBUG] handleAddedConnections - plugin '%s' is still starting - added connections will be sent when it has started", p)
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
		// if the plugin is still starting it has no reattach config yet - treat it as not running
		// (the plugin was started with the connection configs read before this change -
		// startPlugin sends it the changes made since, once it has started)
		if runningPlugin.isStarting() {
			log.Printf("[DEBUG] handleDeletedConnections - plugin '%s' is still starting - deleted connections will be sent when it has started", p)
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
