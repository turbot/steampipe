package pluginmanager_service

import (
	"context"
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

	for _, limitersForPlugin := range m.pluginLimiters {
		for _, l := range limitersForPlugin {
			queries = append(queries, rate_limiters.GetPopulateRateLimiterSql(l))
		}
	}

	for _, limitersForPlugin := range m.userLimiters {
		for _, l := range limitersForPlugin {
			queries = append(queries, rate_limiters.GetPopulateRateLimiterSql(l))
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
