package connection

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
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
	// FetchRateLimiterDefsForConnections contains all connections we need to update the rate limiters defs for
	// in fact we only need to query ask each plugin for the defs so
	// build lookup of the connectionPlugins needed, keyed by plugin name
	connectionPluginsToReladDefs := make(map[string]*steampipeconfig.ConnectionPlugin)
	for plugin := range s.connectionUpdates.FetchRateLimiterDefsForPlugins {
		// find a connection plugin for this plugin
		// // annoying as we key the loaded conne
		for _,
		if connectionPlugin := s.connectionUpdates.ConnectionPlugins[connection]; connectionPlugin != nil {
			connectionPluginsToReladDefs[connectionPlugin.PluginShortName] = connectionPlugin
		}
	}

	var errors []error
	var res = make(map[string]LimiterMap)
	for pluginShortName, connectionPlugin := range connectionPluginsToReladDefs {
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

			m[l.Name] = r
		}
		res[pluginShortName] = m
	}
	return res, nil
}
