package pluginmanager_service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/connection"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"log"
)

func (m *PluginManager) ShouldFetchRateLimiterDefs() bool {
	return m.pluginLimiters == nil
}

// GetPluginExemplarConnections returns a map of keyed by plugin short name with the value an exemplar connection
func (m *PluginManager) GetPluginExemplarConnections() map[string]string {
	res := make(map[string]string)
	for _, c := range m.connectionConfigMap {
		res[c.PluginShortName] = c.Connection
	}
	return res
}

// HandlePluginLimiterChanges responds to changes in the plugin rate limiter defintions
// update the stored limiters, refrresh the rate limiter table and call `setRateLimiters`
// for all plugins with changed limiters
func (m *PluginManager) HandlePluginLimiterChanges(newLimiters map[string]connection.LimiterMap) error {
	if m.pluginLimiters == nil {
		// this must be the first time we have poplkated them
		m.pluginLimiters = make(map[string]connection.LimiterMap)
	}
	for plugin, limitersForPlugin := range newLimiters {
		m.pluginLimiters[plugin] = limitersForPlugin
	}

	// update the rate_limiters table
	if err := m.refreshRateLimiterTable(context.Background()); err != nil {
		log.Println("[WARN] could not refresh rate limiter table", err)
	}
	return nil
}

// respond to changes in the HCL rate limiter config
// update the stored limiters, refrresh the rate limiter table and call `setRateLimiters`
// for all plugins with changed limiters
func (m *PluginManager) handleUserLimiterChanges(newLimiters connection.LimiterMap) error {
	newLimiterPluginMap := newLimiters.ToPluginMap()

	pluginsWithChangedLimiters := m.getPluginsWithChangedLimiters(newLimiterPluginMap)

	if len(pluginsWithChangedLimiters) == 0 {
		return nil
	}

	// update stored limiters to the new map
	m.userLimiters = newLimiterPluginMap

	// update the rate_limiters table
	if err := m.refreshRateLimiterTable(context.Background()); err != nil {
		log.Println("[WARN] could not refresh rate limiter table", err)
	}

	// now update the plugins - call setRateLimiters for any plugin witrh updated user limiters
	for p := range pluginsWithChangedLimiters {
		if err := m.setRateLimitersForPlugin(p); err != nil {
			return err
		}
	}

	return nil
}

func (m *PluginManager) setRateLimitersForPlugin(pluginShortName string) error {
	// get running plugin for this plugin
	// if plugin is not running we have nothing to do
	longName, ok := m.pluginShortToLongNameMap[pluginShortName]
	if !ok {
		log.Printf("[INFO] handleUserLimiterChanges: plugin %s is not currently running - ignoring", pluginShortName)
		return nil
	}
	runningPlugin, ok := m.runningPluginMap[longName]
	if !ok {
		log.Printf("[INFO] handleUserLimiterChanges: plugin %s is not currently running - ignoring", pluginShortName)
		return nil
	}
	if !runningPlugin.reattach.SupportedOperations.RateLimiters {
		log.Printf("[INFO] handleUserLimiterChanges: plugin %s does not support setting rate limit - ignoring", pluginShortName)
		return nil
	}

	pluginClient, err := sdkgrpc.NewPluginClient(runningPlugin.client, longName)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to create a plugin client when updating the rate limiter for plugin '%s'", longName)
	}

	if err := m.setRateLimiters(pluginShortName, pluginClient); err != nil {
		return sperr.WrapWithMessage(err, "failed to update rate limiters for plugin '%s'", longName)
	}
	return nil
}

func (m *PluginManager) getPluginsWithChangedLimiters(newLimiters map[string]connection.LimiterMap) map[string]struct{} {
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
				pluginLimiter.Status = modconfig.LimiterStatusOverriden
			} else {
				pluginLimiter.Status = modconfig.LimiterStatusActive
			}
		}
	}
}

func (m *PluginManager) getUserDefinedLimitersForPlugin(plugin string) connection.LimiterMap {
	log.Printf("[WARN] plugin %s", plugin)
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

	rows, err := m.pool.Query(ctx, fmt.Sprintf("SELECT * FROM %s.%s WHERE source=$1", constants.InternalSchema, constants.RateLimiterDefinitionTable), modconfig.LimiterSourcePlugin)
	if err != nil {
		// TODO KAI if this is a table not found error, do not return error
		return err
	}
	defer rows.Close()

	rateLimiters, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[modconfig.RateLimiter])
	if err != nil {
		return err
	}

	// ok so populate pluginLimiters
	m.pluginLimiters = make(map[string]connection.LimiterMap)
	for _, r := range rateLimiters {
		limitersForPlugin := m.pluginLimiters[r.Plugin]
		if limitersForPlugin == nil {
			limitersForPlugin = make(connection.LimiterMap)
		}
		limitersForPlugin[r.Name] = r
		m.pluginLimiters[r.Plugin] = limitersForPlugin
	}

	// then (re)write the steampipe_rate_limiter table
	// this is to ensure we include any updates made to the rate limiter config since the last execution)
	return m.refreshRateLimiterTable(ctx)

}

func (m *PluginManager) rateLimiterTableExists(ctx context.Context) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS (
    SELECT FROM 
        pg_tables
    WHERE 
        schemaname = '%s' AND 
        tablename  = '%s'
    );`, constants.InternalSchema, constants.RateLimiterDefinitionTable)

	row := m.pool.QueryRow(ctx, query)
	var exists bool
	err := row.Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil
}
