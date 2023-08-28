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

func (s *refreshConnectionState) ensureRateLimiterTable(ctx context.Context) error {
	// if the rate limiter table exists, nothing to do
	exists, err := s.rateLimiterTableExists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// so the table does not exist
	// all we need to do is populate FetchRateLimiterDefsForConnections with a list of all connecitons
	// - then the connection updates will start all plugins and when we refresh the rate limiter table
	// we will fully populate it
	s.fetchRateLimiterDefsForConnectionNames = maps.Keys(s.pluginManager.GetConnectionConfig())

	return nil
}

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
	connectionPlugins := make(map[string]*steampipeconfig.ConnectionPlugin)
	for connection := range s.connectionUpdates.FetchRateLimiterDefsForConnections {
		conectionPlugin := s.connectionUpdates.ConnectionPlugins[connection]
		connectionPlugins[conectionPlugin.PluginName] = conectionPlugin
	}
	var errors []error
	var res = make(map[string]LimiterMap)
	for pluginName, connectionPlugin := range connectionPlugins {
		rateLimiterResp, err := connectionPlugin.PluginClient.GetRateLimiters(&proto.GetRateLimitersRequest{})
		if err != nil {
			return nil, err
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
		res[pluginName] = m
	}
	return res, nil
}
