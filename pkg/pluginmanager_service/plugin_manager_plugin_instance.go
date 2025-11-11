package pluginmanager_service

import (
	"context"

	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe/v2/pkg/connection"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
	"golang.org/x/exp/maps"
)

func (m *PluginManager) handlePluginInstanceChanges(ctx context.Context, newPlugins connection.PluginMap) error {
	if maps.EqualFunc(m.plugins, newPlugins, func(l *plugin.Plugin, r *plugin.Plugin) bool {
		return l.Equals(r)
	}) {
		return nil
	}

	// now determine whether there are any new or deleted connections
	//addedConnections, deletedConnections, changedConnections := m.plugins.Diff(newPlugins)

	//m.handleDeletedPlugins(deletedConnections, requestMap)
	//
	//m.handleAddedPlugins(addedConnections, requestMap)
	//m.handleUpdatedPlugins(changedConnections, requestMap)

	// update connectionConfigMap
	m.plugins = newPlugins

	// if pool is nil, we're in a test environment or the plugin manager hasn't been fully initialized
	// in this case, we can't repopulate the plugin table, so just return early
	if m.pool == nil {
		return nil
	}

	// repopulate the plugin table
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	return db_local.PopulatePluginTable(ctx, conn.Conn())

}
