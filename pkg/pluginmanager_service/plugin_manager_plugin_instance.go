package pluginmanager_service

import (
	"context"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/steampipe/pkg/connection"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"golang.org/x/exp/maps"
)

func (m *PluginManager) handlePluginInstanceChanges(ctx context.Context, newPlugins connection.PluginMap) error {
	if maps.EqualFunc(m.plugins, newPlugins, func(l *modconfig.Plugin, r *modconfig.Plugin) bool {
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

	// repopulate the plugin table
	conn, err := m.pool.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return db_local.PopulatePluginTable(ctx, conn)

}
