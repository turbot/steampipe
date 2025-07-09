package pluginmanager_service

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
	"github.com/turbot/steampipe/v2/pkg/introspection"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
	"golang.org/x/exp/maps"
)

func (m *PluginManager) initialisePluginColumns(ctx context.Context) error {
	if m.shouldBootstrapPluginColumnTable(ctx) {
		return m.bootstrapPluginColumnTable(ctx)
	}
	return nil
}

func (m *PluginManager) shouldBootstrapPluginColumnTable(ctx context.Context) bool {
	pluginColumnTableExists, err := m.tableExists(ctx, constants.InternalSchema, constants.PluginColumnTable)
	if err != nil || !pluginColumnTableExists {
		return true
	}
	// check columns match
	query := fmt.Sprintf(`SELECT column_name
  FROM information_schema.columns
 WHERE table_schema = '%s'
   AND table_name   = '%s'`, constants.InternalSchema, constants.PluginColumnTable)

	rows, err := m.pool.Query(ctx, query)
	if err != nil {
		return true
	}
	defer rows.Close()
	var columns []string
	// Iterate through the rows
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return true
		}
		columns = append(columns, s)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return true
	}

	expectedColumns := []string{"plugin_", "table_name", "name", "type", " description", "list_config", "get_config", "hydrate_name", "default_value"}
	return !slices.Equal(columns, expectedColumns)
}

func (m *PluginManager) bootstrapPluginColumnTable(ctx context.Context) error {
	schemas, err := m.loadPluginSchemas(m.getPluginExemplarConnections())
	if err != nil {
		log.Printf("[WARN] loadPluginSchemas error: %s", err.Error())
		return err
	}

	if err := m.createPluginColumnsTable(ctx); err != nil {
		log.Printf("[WARN] createPluginColumnsTable error: %s", err.Error())
		return err
	}
	// now populate the table
	log.Printf("[INFO] bootstrapPluginColumnTable loaded schema for plugins:  %s", strings.Join(maps.Keys(schemas), ","))

	return m.populatePluginColumnsTable(ctx, schemas)
}

func (m *PluginManager) createPluginColumnsTable(ctx context.Context) error {
	queries := []db_common.QueryWithArgs{
		introspection.GetPluginColumnTableDropSql(),
		introspection.GetPluginColumnTableCreateSql(),
		introspection.GetPluginColumnTableGrantSql(),
	}

	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = db_local.ExecuteSqlWithArgsInTransaction(ctx, conn.Conn(), queries...)
	return err
}

func (m *PluginManager) populatePluginColumnsTable(ctx context.Context, schemas map[string]*proto.Schema) error {
	if len(schemas) == 0 {
		log.Printf("[INFO] populatePluginColumnsTable : no updates to plugin_columns table")
		return nil
	}
	log.Printf("[INFO] populating plugin_columns table for plugins %s", strings.Join(maps.Keys(schemas), ","))
	var queries []db_common.QueryWithArgs
	for plugin, schema := range schemas {
		// drop entries for this plugin
		queries = append(queries, introspection.GetPluginColumnTableDeletePluginSql(plugin))

		// NOTE: we do not support dynamic plugins
		if schema.Mode == sdkplugin.SchemaModeDynamic {
			continue
		}
		pluginQueries, err := introspection.GetPluginColumnTablePopulateSqlForPlugin(plugin, schema.Schema)
		if err != nil {
			return err
		}
		queries = append(queries, pluginQueries...)
	}

	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = db_local.ExecuteSqlWithArgsInTransaction(ctx, conn.Conn(), queries...)
	return err
}

func (m *PluginManager) removePluginsFromPluginColumnsTable(ctx context.Context, plugins []string) error {
	if len(plugins) == 0 {
		return nil
	}

	var queries []db_common.QueryWithArgs

	for _, plugin := range plugins {
		// drop entries for this plugin
		queries = append(queries, introspection.GetPluginColumnTableDeletePluginSql(plugin))
	}

	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = db_local.ExecuteSqlWithArgsInTransaction(ctx, conn.Conn(), queries...)
	return err
}

// load the schemas for the given plugin connections
func (m *PluginManager) loadPluginSchemas(pluginConnectionMap map[string]string) (map[string]*proto.Schema, error) {
	// build Get request
	req := &pb.GetRequest{
		Connections: maps.Values(pluginConnectionMap),
	}
	plugins, err := m.Get(req)
	if err != nil {
		return nil, err
	}
	var res = make(map[string]*proto.Schema)

	// ok so now we have all necessary plugin reattach configs - fetch the schemas

	for _, reattach := range plugins.ReattachMap {
		// attach to the plugin process
		pluginClient, err := sdkgrpc.NewPluginClientFromReattach(reattach.Convert(), reattach.Plugin)
		if err != nil {
			log.Printf("[WARN] failed to attach to plugin '%s' - pid %d: %s",
				reattach.Plugin, reattach.Pid, err)
			return nil, err
		}

		schemaResp, err := pluginClient.GetSchema(reattach.Connections[0])
		if err != nil {
			return nil, err
		}

		res[reattach.Plugin] = schemaResp
	}

	return res, nil
}

func (m *PluginManager) UpdatePluginColumnsTable(ctx context.Context, update map[string]*proto.Schema, delete []string) error {
	if err := m.removePluginsFromPluginColumnsTable(ctx, delete); err != nil {
		return err
	}
	return m.populatePluginColumnsTable(ctx, update)
}
