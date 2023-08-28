package pluginmanager_service

import (
	"context"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/rate_limiters"
)

func (m *PluginManager) refreshRateLimiterTable(ctx context.Context) error {
	queries := []db_common.QueryWithArgs{
		rate_limiters.DropRateLimiterTable(),
		rate_limiters.CreateRateLimiterTable(),
		rate_limiters.GrantsOnRateLimiterTable(),
	}
	// build a resolved set of rate limiter def from the plugin and user defined rate limiters
	var resolvedLimiters = m.resolveRateLimiterDefs()
	for _, resolvedLimiter := range resolvedLimiters {
		queries = append(queries, rate_limiters.GetPopulateRateLimiterSql(resolvedLimiter))
	}

	conn, err := db_local.CreateLocalDbConnection(ctx, &db_local.CreateDbOptions{
		Username: constants.DatabaseSuperUser,
	})
	if err != nil {
		return err
	}
	_, err = db_local.ExecuteSqlWithArgsInTransaction(ctx, conn, queries...)
	return err
}
