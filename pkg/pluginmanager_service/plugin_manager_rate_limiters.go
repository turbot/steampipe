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

// respond to changes in the HCL rate limiter config
// update the stored limiters, refrresh the rate limiter table and call `setRateLimiters`
// for all plugins with changed limiters
func (m *PluginManager) HandlePluginLimiterChanges(newLimiters map[string]connection.LimiterMap) error {
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

	// now update the plugins
	for p := range pluginsWithChangedLimiters {
		// get running plugin for this plugin
		// if plugin is not running we have nothing to do
		longName, ok := m.pluginShortToLongNameMap[p]
		if !ok {
			log.Printf("[INFO] handleUserLimiterChanges: plugin %s is not currently running - ignoring", p)
			continue
		}
		runningPlugin, ok := m.runningPluginMap[longName]
		if !ok {
			log.Printf("[INFO] handleUserLimiterChanges: plugin %s is not currently running - ignoring", p)
			continue
		}
		if !runningPlugin.reattach.SupportedOperations.RateLimiters {
			log.Printf("[INFO] handleUserLimiterChanges: plugin %s does not support setting rate limit - ignoring", p)
			continue
		}

		pluginClient, err := sdkgrpc.NewPluginClient(runningPlugin.client, longName)
		if err != nil {
			return sperr.WrapWithMessage(err, "failed to create a plugin client when updating the rate limiter for plugin '%s'", longName)
		}

		if err := m.setRateLimiters(p, pluginClient); err != nil {
			return sperr.WrapWithMessage(err, "failed to update rate limiters for plugin '%s'", longName)
		}
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

func (m *PluginManager) resolveRateLimiterDefs() (res []*modconfig.ResolvedRateLimiter) {
	// add all plugin limiters
	for plugin, pluginDefinedLimiters := range m.pluginLimiters {
		// are there any user defined limiters for this plugin
		userDefinedLimiters := m.getUserDefinedLimitersForPlugin(plugin)
		log.Printf("[WARN] %v", userDefinedLimiters)
		for name, pluginLimiter := range pluginDefinedLimiters {
			resolvedLimiter := modconfig.NewResolvedRateLimiter(pluginLimiter, modconfig.LimiterStatusActive, modconfig.LimiterSourcePlugin)

			// is there a user override - if set set status to overriden
			if _, isOverriden := userDefinedLimiters[name]; isOverriden {
				resolvedLimiter.Status = modconfig.LimiterStatusOverriden
			}

			res = append(res, resolvedLimiter)
		}

		// add all user defined limiters
		for _, userLimiter := range userDefinedLimiters {
			resolvedLimiter := modconfig.NewResolvedRateLimiter(userLimiter, modconfig.LimiterStatusActive, modconfig.LimiterSourceConfig)
			res = append(res, resolvedLimiter)
		}
	}
	return res
}

func (m *PluginManager) getUserDefinedLimitersForPlugin(plugin string) connection.LimiterMap {
	log.Printf("[WARN] plugin %s", plugin)
	userDefinedLimiters := m.userLimiters[plugin]
	if userDefinedLimiters == nil {
		userDefinedLimiters = make(connection.LimiterMap)
	}
	return userDefinedLimiters
}

func (m *PluginManager) populatePluginRateLimiterDefs(ctx context.Context) (e error) {

	// if the rate limiter table exists, nothing to do
	exists, err := m.rateLimiterTableExists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	defer func() {
		// this function uses reflection to extract and convert values
		// we need to be able to recover from panics while using reflection
		if r := recover(); r != nil {
			e = sperr.ToError(r, sperr.WithMessage("error loading server settings"))
		}
	}()
	rows, err := conn.Query(ctx, fmt.Sprintf("SELECT * FROM %s.%s", constants.InternalSchema, constants.RateLimiterDefinitionTable))
	if err != nil {
		return err
	}
	defer rows.Close()

	rateLimiters, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[modconfig.ResolvedRateLimiter])

	if err != nil {
		return err
	}

	// ok so populate pluginLimiters
	m.pluginLimiters = make(map[string]connection.LimiterMap)
	for _, r := range rateLimiters {
		if r.Source == modconfig.LimiterSourcePlugin {
			limitersForPlugin := m.pluginLimiters[r.Plugin]
			if limitersForPlugin == nil {
				limitersForPlugin = make(connection.LimiterMap)
			}
			limitersForPlugin[r.Name] = r.Limiter()
			m.pluginLimiters[r.Plugin] = limitersForPlugin
		}

	}
	return nil

}

func (s *PluginManager) rateLimiterTableExists(ctx context.Context) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS (
    SELECT FROM 
        pg_tables
    WHERE 
        schemaname = '%s' AND 
        tablename  = '%s'
    );`, constants.InternalSchema, constants.RateLimiterDefinitionTable)

	row := s.pool.QueryRow(ctx, query)
	var exists bool
	err := row.Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil
}
