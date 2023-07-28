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
	for _, limiter := range m.limiters {
		queries = append(queries, rate_limiters.GetPopulateRateLimiterSql(limiter))
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
