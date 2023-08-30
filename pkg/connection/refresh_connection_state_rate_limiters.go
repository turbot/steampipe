package connection

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"golang.org/x/exp/maps"
)

func (s *refreshConnectionState) rateLimiterTableExists(ctx context.Context) (bool, error) {
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

func (s *refreshConnectionState) reloadPluginRateLimiters() (map[string]LimiterMap, error) {
	// build lookup of the connectionPlugins we need to fetch rate limiter defs for
	var connectionPluginsToReloadDefs []*steampipeconfig.ConnectionPlugin

	// build map of the connection plugins already started, keyed by plugin short name
	connectionPluginsByPlugin := make(map[string]*steampipeconfig.ConnectionPlugin)
	for _, c := range s.connectionUpdates.ConnectionPlugins {
		// potentially overwire so we only store the final connection per plugin
		connectionPluginsByPlugin[c.PluginShortName] = c
	}

	var missingPlugins []string
	for plugin := range s.connectionUpdates.FetchRateLimiterDefsForPlugins {
		if connectionPlugin, started := connectionPluginsByPlugin[plugin]; started {
			connectionPluginsToReloadDefs = append(connectionPluginsToReloadDefs, connectionPlugin)
		} else {
			missingPlugins = append(missingPlugins, plugin)
		}
	}
	if len(missingPlugins) > 0 {
		missingConnectionPlugins, res := s.startPlugins(missingPlugins)
		if res.Error != nil {
			// TODO just warn?
			return nil, res.Error
		}
		connectionPluginsToReloadDefs = append(connectionPluginsToReloadDefs, maps.Values(missingConnectionPlugins)...)
	}

	// ok so now we have all necessary connection plugins - fetch the rate limiter defs
	var errors []error
	var res = make(map[string]LimiterMap)
	for _, connectionPlugin := range connectionPluginsToReloadDefs {
		if !connectionPlugin.SupportedOperations.RateLimiters {
			continue
		}
		rateLimiterResp, err := connectionPlugin.PluginClient.GetRateLimiters(&proto.GetRateLimitersRequest{})
		if err != nil {
			return nil, err
		}
		if rateLimiterResp == nil || rateLimiterResp.Definitions == nil {
			continue
		}
		m := make(LimiterMap)
		for _, l := range rateLimiterResp.Definitions {
			r, err := modconfig.RateLimiterFromProto(l)
			if err != nil {
				errors = append(errors, sperr.WrapWithMessage(err, "failed to create rate limiter %s from plugin definition", err))
				continue
			}
			// populate the plugin name
			r.Plugin = connectionPlugin.PluginShortName
			// set plugin as source
			r.Source = modconfig.LimiterSourcePlugin
			// derfaulty status to active
			r.Status = modconfig.LimiterStatusActive
			// add to map
			m[l.Name] = r
		}
		// store back
		res[connectionPlugin.PluginShortName] = m
	}
	return res, nil
}

func (s *refreshConnectionState) startPlugins(plugins []string) (map[string]*steampipeconfig.ConnectionPlugin, *steampipeconfig.RefreshConnectionResult) {
	var exemplarConnections []string
	connectionConfig := s.pluginManager.GetConnectionConfig()
	// for each plugin we need to find an exemplat connection
	for _, p := range plugins {
		for _, c := range connectionConfig {
			if c.PluginShortName == p {
				exemplarConnections = append(exemplarConnections, c.Connection)
			}
		}
	}
	return steampipeconfig.CreateConnectionPlugins(exemplarConnections)

}
