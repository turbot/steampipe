package pluginmanager_service

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/pipe-fittings/v2/plugin"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/connection"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
	"github.com/turbot/steampipe/v2/pkg/introspection"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
	"golang.org/x/exp/maps"
)

func (m *PluginManager) ShouldFetchRateLimiterDefs() bool {
	return m.pluginLimiters == nil
}

// HandlePluginLimiterChanges responds to changes in the plugin rate limiter definitions
// update the stored limiters, refrresh the rate limiter table and call `setRateLimiters`
// for all plugins with changed limiters
func (m *PluginManager) HandlePluginLimiterChanges(newLimiters connection.PluginLimiterMap) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	if m.pluginLimiters == nil {
		// this must be the first time we have populated them
		m.pluginLimiters = make(connection.PluginLimiterMap)
	}
	for plugin, limitersForPlugin := range newLimiters {
		m.pluginLimiters[plugin] = limitersForPlugin
	}

	// update the steampipe_plugin_limiters table
	// NOTE: we hold m.mut lock, so call internal version
	if err := m.refreshRateLimiterTableInternal(context.Background()); err != nil {
		log.Println("[WARN] could not refresh rate limiter table", err)
	}
	return nil
}

func (m *PluginManager) refreshRateLimiterTable(ctx context.Context) error {
	m.mut.Lock()
	defer m.mut.Unlock()
	return m.refreshRateLimiterTableInternal(ctx)
}

func (m *PluginManager) refreshRateLimiterTableInternal(ctx context.Context) error {
	// NOTE: caller must hold m.mut lock

	// if we have not yet populated the rate limiter table, do nothing
	if m.pluginLimiters == nil {
		return nil
	}

	// if the pool is nil, we cannot refresh the table
	if m.pool == nil {
		return nil
	}

	// update the status of the plugin rate limiters (determine which are overriden and set state accordingly)
	m.updateRateLimiterStatusInternal()

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

	// NOTE: no lock needed here, caller already holds m.mut
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
	log.Printf("[DEBUG] handleUserLimiterChanges: start")
	limiterPluginMap := plugins.ToPluginLimiterMap()
	log.Printf("[DEBUG] handleUserLimiterChanges: got limiter plugin map")
	// NOTE: caller (OnConnectionConfigChanged) already holds m.mut lock, so use internal version
	pluginsWithChangedLimiters := m.getPluginsWithChangedLimitersInternal(limiterPluginMap)
	log.Printf("[DEBUG] handleUserLimiterChanges: found %d plugins with changed limiters", len(pluginsWithChangedLimiters))

	if len(pluginsWithChangedLimiters) == 0 {
		log.Printf("[DEBUG] handleUserLimiterChanges: no changes, returning")
		return nil
	}

	// update stored limiters to the new map
	// NOTE: caller (OnConnectionConfigChanged) already holds m.mut lock, so we don't lock here
	log.Printf("[DEBUG] handleUserLimiterChanges: updating user limiters")
	m.userLimiters = limiterPluginMap

	// update the steampipe_plugin_limiters table
	// NOTE: caller already holds m.mut lock, so call internal version
	log.Printf("[DEBUG] handleUserLimiterChanges: calling refreshRateLimiterTableInternal")
	if err := m.refreshRateLimiterTableInternal(context.Background()); err != nil {
		log.Println("[WARN] could not refresh rate limiter table", err)
	}
	log.Printf("[DEBUG] handleUserLimiterChanges: refreshRateLimiterTableInternal complete")

	// now update the plugins - call setRateLimiters for any plugin with updated user limiters
	log.Printf("[DEBUG] handleUserLimiterChanges: setting rate limiters for plugins")
	for p := range pluginsWithChangedLimiters {
		log.Printf("[DEBUG] handleUserLimiterChanges: calling setRateLimitersForPlugin for %s", p)
		if err := m.setRateLimitersForPlugin(p); err != nil {
			return err
		}
		log.Printf("[DEBUG] handleUserLimiterChanges: setRateLimitersForPlugin complete for %s", p)
	}

	log.Printf("[DEBUG] handleUserLimiterChanges: complete")
	return nil
}

func (m *PluginManager) setRateLimitersForPlugin(pluginShortName string) error {
	// get running plugin for this plugin
	imageRef := ociinstaller.NewImageRef(pluginShortName).DisplayImageRef()

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

	// NOTE: caller (handleUserLimiterChanges via OnConnectionConfigChanged) already holds m.mut lock
	if err := m.setRateLimitersInternal(pluginShortName, pluginClient); err != nil {
		return sperr.WrapWithMessage(err, "failed to update rate limiters for plugin '%s'", imageRef)
	}
	return nil
}

func (m *PluginManager) getPluginsWithChangedLimiters(newLimiters connection.PluginLimiterMap) map[string]struct{} {
	m.mut.RLock()
	defer m.mut.RUnlock()
	return m.getPluginsWithChangedLimitersInternal(newLimiters)
}

func (m *PluginManager) getPluginsWithChangedLimitersInternal(newLimiters connection.PluginLimiterMap) map[string]struct{} {
	// NOTE: caller must hold m.mut lock (at least RLock)
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
	m.mut.Lock()
	defer m.mut.Unlock()
	m.updateRateLimiterStatusInternal()
}

func (m *PluginManager) updateRateLimiterStatusInternal() {
	// NOTE: caller must hold m.mut lock
	// iterate through limiters for each plug
	for p, pluginDefinedLimiters := range m.pluginLimiters {
		// get user limiters for this plugin (already holding lock, so call internal version)
		userDefinedLimiters := m.getUserDefinedLimitersForPluginInternal(p)

		// is there a user override? - if so set status to overriden
		for name, pluginLimiter := range pluginDefinedLimiters {
			_, isOverriden := userDefinedLimiters[name]
			if isOverriden {
				pluginLimiter.Status = plugin.LimiterStatusOverridden
			} else {
				pluginLimiter.Status = plugin.LimiterStatusActive
			}
		}
	}
}

func (m *PluginManager) getUserDefinedLimitersForPlugin(plugin string) connection.LimiterMap {
	m.mut.RLock()
	defer m.mut.RUnlock()
	return m.getUserDefinedLimitersForPluginInternal(plugin)
}

// getUserDefinedLimitersForPluginInternal returns user-defined limiters for a plugin
// WITHOUT acquiring the lock - caller must hold the lock
func (m *PluginManager) getUserDefinedLimitersForPluginInternal(plugin string) connection.LimiterMap {
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

func (m *PluginManager) loadRateLimitersFromTable(ctx context.Context) ([]*plugin.RateLimiter, error) {
	rows, err := m.pool.Query(ctx, fmt.Sprintf("SELECT * FROM %s.%s", constants.InternalSchema, constants.RateLimiterDefinitionTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rateLimiters, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[plugin.RateLimiter])
	if err != nil {
		return nil, err
	}
	// convert to pointer array
	pRateLimiters := make([]*plugin.RateLimiter, len(rateLimiters))
	for i, r := range rateLimiters {
		// copy into loop var
		rateLimiter := r
		pRateLimiters[i] = &rateLimiter
	}
	return pRateLimiters, nil
}

func (m *PluginManager) getUserAndPluginLimitersFromTableResult(rateLimiters []*plugin.RateLimiter) (connection.PluginLimiterMap, connection.PluginLimiterMap) {
	pluginLimiters := make(connection.PluginLimiterMap)
	userLimiters := make(connection.PluginLimiterMap)
	for _, r := range rateLimiters {
		if r.Source == plugin.LimiterSourcePlugin {
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
			r, err := RateLimiterFromProto(l, reattach.Plugin, pluginInstance)
			if err != nil {
				errors = append(errors, sperr.WrapWithMessage(err, "failed to create rate limiter %s from plugin definition", err))
				continue
			}

			// set plugin as source
			r.Source = plugin.LimiterSourcePlugin
			// default status to active
			r.Status = plugin.LimiterStatusActive
			// add to map
			limitersForPlugin[l.Name] = r
		}
		// store back
		res[reattach.Plugin] = limitersForPlugin
	}

	if len(errors) > 0 {
		return nil, error_helpers.CombineErrors(errors...)
	}

	return res, nil
}
