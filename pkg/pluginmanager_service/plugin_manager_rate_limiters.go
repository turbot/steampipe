package pluginmanager_service

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/connection"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/introspection"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	pb "github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/proto"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"golang.org/x/exp/maps"
)

func (m *PluginManager) ShouldFetchRateLimiterDefs() bool {
	return m.pluginLimiters == nil
}

// HandlePluginLimiterChanges responds to changes in the plugin rate limiter definitions
// update the stored limiters, refrresh the rate limiter table and call `setRateLimiters`
// for all plugins with changed limiters
func (m *PluginManager) HandlePluginLimiterChanges(newLimiters connection.PluginLimiterMap) error {
	if m.pluginLimiters == nil {
		// this must be the first time we have populated them
		m.pluginLimiters = make(connection.PluginLimiterMap)
	}
	for plugin, limitersForPlugin := range newLimiters {
		m.pluginLimiters[plugin] = limitersForPlugin
	}

	// update the steampipe_plugin_limiters table
	if err := m.refreshRateLimiterTable(context.Background()); err != nil {
		log.Println("[WARN] could not refresh rate limiter table", err)
	}
	return nil
}

func (m *PluginManager) refreshRateLimiterTable(ctx context.Context) error {
	// if we have not yet populated the rate limiter table, do nothing
	if m.pluginLimiters == nil {
		return nil
	}

	// update the status of the plugin rate limiters (determine which are overriden and set state accordingly)
	m.updateRateLimiterStatus()

	queries := []db_common.QueryWithArgs{
		introspection.GetRateLimiterTableDropSql(),
		introspection.GetRateLimiterTableCreateSql(),
		introspection.GetRateLimiterTableGrantSql(),
	}

	for _, limitersForPlugin := range m.pluginLimiters {
		for _, l := range limitersForPlugin {
			queries = append(queries, introspection.GetRateLimiterTablePopulateSql(l))
		}
	}

	for _, limitersForPlugin := range m.userLimiters {
		for _, l := range limitersForPlugin {
			queries = append(queries, introspection.GetRateLimiterTablePopulateSql(l))
		}
	}

	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = db_local.ExecuteSqlWithArgsInTransaction(ctx, conn.Conn(), queries...)
	return err
}

// respond to changes in the HCL rate limiter config
// update the stored limiters, refresh the rate limiter table and call `setRateLimiters`
// for all plugins with changed limiters
func (m *PluginManager) handleUserLimiterChanges(_ context.Context, plugins connection.PluginMap) error {
	limiterPluginMap := plugins.ToPluginLimiterMap()
	pluginsWithChangedLimiters := m.getPluginsWithChangedLimiters(limiterPluginMap)

	if len(pluginsWithChangedLimiters) == 0 {
		return nil
	}

	// update stored limiters to the new map
	m.userLimiters = limiterPluginMap

	// update the steampipe_plugin_limiters table
	if err := m.refreshRateLimiterTable(context.Background()); err != nil {
		log.Println("[WARN] could not refresh rate limiter table", err)
	}

	// now update the plugins - call setRateLimiters for any plugin with updated user limiters
	for p := range pluginsWithChangedLimiters {
		if err := m.setRateLimitersForPlugin(p); err != nil {
			return err
		}
	}

	return nil
}

func (m *PluginManager) setRateLimitersForPlugin(pluginShortName string) error {
	// get running plugin for this plugin
	imageRef := ociinstaller.NewSteampipeImageRef(pluginShortName).DisplayImageRef()

	runningPlugin, ok := m.runningPluginMap[imageRef]
	if !ok {
		log.Printf("[INFO] handleUserLimiterChanges: plugin %s is not currently running - ignoring", pluginShortName)
		return nil
	}
	if !runningPlugin.reattach.SupportedOperations.RateLimiters {
		log.Printf("[INFO] handleUserLimiterChanges: plugin %s does not support setting rate limit - ignoring", pluginShortName)
		return nil
	}

	pluginClient, err := sdkgrpc.NewPluginClient(runningPlugin.client, imageRef)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to create a plugin client when updating the rate limiter for plugin '%s'", imageRef)
	}

	if err := m.setRateLimiters(pluginShortName, pluginClient); err != nil {
		return sperr.WrapWithMessage(err, "failed to update rate limiters for plugin '%s'", imageRef)
	}
	return nil
}

func (m *PluginManager) getPluginsWithChangedLimiters(newLimiters connection.PluginLimiterMap) map[string]struct{} {
	var pluginsWithChangedLimiters = make(map[string]struct{})

	for plugin, limitersForPlugin := range m.userLimiters {
		newLimitersForPlugin := newLimiters[plugin]
		if !limitersForPlugin.Equals(newLimitersForPlugin) {
			pluginsWithChangedLimiters[plugin] = struct{}{}
		}
	}
	// look for plugins did not have limiters before
	for plugin := range newLimiters {
		_, pluginHasLimiters := m.userLimiters[plugin]
		if !pluginHasLimiters {
			pluginsWithChangedLimiters[plugin] = struct{}{}
		}
	}
	return pluginsWithChangedLimiters
}

func (m *PluginManager) updateRateLimiterStatus() {
	// iterate through limiters for each plug
	for plugin, pluginDefinedLimiters := range m.pluginLimiters {
		// get user limiters for this plugin
		userDefinedLimiters := m.getUserDefinedLimitersForPlugin(plugin)

		// is there a user override? - if so set status to overriden
		for name, pluginLimiter := range pluginDefinedLimiters {
			_, isOverriden := userDefinedLimiters[name]
			if isOverriden {
				pluginLimiter.Status = modconfig.LimiterStatusOverridden
			} else {
				pluginLimiter.Status = modconfig.LimiterStatusActive
			}
		}
	}
}

func (m *PluginManager) getUserDefinedLimitersForPlugin(plugin string) connection.LimiterMap {
	userDefinedLimiters := m.userLimiters[plugin]
	if userDefinedLimiters == nil {
		userDefinedLimiters = make(connection.LimiterMap)
	}
	return userDefinedLimiters
}

func (m *PluginManager) initialiseRateLimiterDefs(ctx context.Context) (e error) {
	defer func() {
		// this function uses reflection to extract and convert values
		// we need to be able to recover from panics while using reflection
		if r := recover(); r != nil {
			e = sperr.ToError(r, sperr.WithMessage("error loading rate limiter definitions"))
		}
	}()

	rateLimiterTableExists, err := m.tableExists(ctx, constants.InternalSchema, constants.RateLimiterDefinitionTable)
	if err != nil {
		return err
	}

	if !rateLimiterTableExists {
		return m.bootstrapRateLimiterTable(ctx)
	}

	rateLimiters, err := m.loadRateLimitersFromTable(ctx)
	if err != nil {
		return err
	}

	// split the table result into plugin and user limiters
	pluginLimiters, previousUserLimiters := m.getUserAndPluginLimitersFromTableResult(rateLimiters)
	// store the plugin limiters
	m.pluginLimiters = pluginLimiters

	if previousUserLimiters.Equals(m.userLimiters) {
		return nil
	}
	// if the user limiter in the table are different from the current user listeners, the config must have changed
	// since we last ran - call refreshRateLimiterTable to (re)write the steampipe_rate_limiter table
	return m.refreshRateLimiterTable(ctx)
}

func (m *PluginManager) bootstrapRateLimiterTable(ctx context.Context) error {
	pluginLimiters, err := m.LoadPluginRateLimiters(m.getPluginExemplarConnections())
	if err != nil {
		return err
	}
	m.pluginLimiters = pluginLimiters
	// now populate the table
	return m.refreshRateLimiterTable(ctx)
}

func (m *PluginManager) loadRateLimitersFromTable(ctx context.Context) ([]*modconfig.RateLimiter, error) {
	rows, err := m.pool.Query(ctx, fmt.Sprintf("SELECT * FROM %s.%s", constants.InternalSchema, constants.RateLimiterDefinitionTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rateLimiters, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[modconfig.RateLimiter])
	if err != nil {
		return nil, err
	}
	// convert to pointer array
	pRateLimiters := make([]*modconfig.RateLimiter, len(rateLimiters))
	for i, r := range rateLimiters {
		// copy into loop var
		rateLimiter := r
		pRateLimiters[i] = &rateLimiter
	}
	return pRateLimiters, nil
}

func (m *PluginManager) getUserAndPluginLimitersFromTableResult(rateLimiters []*modconfig.RateLimiter) (connection.PluginLimiterMap, connection.PluginLimiterMap) {
	pluginLimiters := make(connection.PluginLimiterMap)
	userLimiters := make(connection.PluginLimiterMap)
	for _, r := range rateLimiters {
		if r.Source == modconfig.LimiterSourcePlugin {
			pluginLimitersForPlugin := pluginLimiters[r.Plugin]
			if pluginLimitersForPlugin == nil {
				pluginLimitersForPlugin = make(connection.LimiterMap)
			}

			pluginLimitersForPlugin[r.Name] = r
			pluginLimiters[r.Plugin] = pluginLimitersForPlugin
		} else {
			userLimitersForPlugin := userLimiters[r.Plugin]
			if userLimitersForPlugin == nil {
				userLimitersForPlugin = make(connection.LimiterMap)
			}
			userLimitersForPlugin[r.Name] = r
			userLimiters[r.Plugin] = userLimitersForPlugin
		}
	}
	return pluginLimiters, userLimiters
}

func (m *PluginManager) LoadPluginRateLimiters(pluginConnectionMap map[string]string) (connection.PluginLimiterMap, error) {
	// build Get request
	req := &pb.GetRequest{
		Connections: maps.Values(pluginConnectionMap),
	}
	resp, err := m.Get(req)
	if err != nil {
		return nil, err
	}

	// ok so now we have all necessary plugin reattach configs - fetch the rate limiter defs
	var errors []error
	var res = make(connection.PluginLimiterMap)
	for pluginInstance, reattach := range resp.ReattachMap {

		if !reattach.SupportedOperations.RateLimiters {
			continue
		}
		// attach to the plugin process
		pluginClient, err := sdkgrpc.NewPluginClientFromReattach(reattach.Convert(), reattach.Plugin)
		if err != nil {
			log.Printf("[WARN] failed to attach to plugin '%s' - pid %d: %s",
				reattach.Plugin, reattach.Pid, err)
			return nil, err
		}
		rateLimiterResp, err := pluginClient.GetRateLimiters(&proto.GetRateLimitersRequest{})
		if err != nil {
			return nil, err
		}
		if rateLimiterResp == nil || rateLimiterResp.Definitions == nil {
			continue
		}

		limitersForPlugin := make(connection.LimiterMap)
		for _, l := range rateLimiterResp.Definitions {
			r, err := modconfig.RateLimiterFromProto(l, reattach.Plugin, pluginInstance)
			if err != nil {
				errors = append(errors, sperr.WrapWithMessage(err, "failed to create rate limiter %s from plugin definition", err))
				continue
			}

			// set plugin as source
			r.Source = modconfig.LimiterSourcePlugin
			// default status to active
			r.Status = modconfig.LimiterStatusActive
			// add to map
			limitersForPlugin[l.Name] = r
		}
		// store back
		res[reattach.Plugin] = limitersForPlugin
	}

	return res, nil
}
