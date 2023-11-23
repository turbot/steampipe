package pluginmanager_service

import (
	"context"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/introspection"
)

func (m *PluginManager) refreshPluginTable(ctx context.Context) error {
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

//
//func (m *PluginManager) refreshRateLimiterTable(ctx context.Context) error {
//	// if we have not yet populated the rate limiter table, do nothing
//	if m.pluginLimiters == nil {
//		return nil
//	}
//
//	// update the status of the plugin rate limiters (determine which are overriden and set state accordingly)
//	m.updateRateLimiterStatus()
//
//	queries := []db_common.QueryWithArgs{
//		introspection.GetRateLimiterTableDropSql(),
//		introspection.GetRateLimiterTableCreateSql(),
//		introspection.GetRateLimiterTableGrantSql(),
//	}
//
//	for _, limitersForPlugin := range m.pluginLimiters {
//		for _, l := range limitersForPlugin {
//			queries = append(queries, introspection.GetRateLimiterTablePopulateSql(l))
//		}
//	}
//
//	for _, limitersForPlugin := range m.userLimiters {
//		for _, l := range limitersForPlugin {
//			queries = append(queries, introspection.GetRateLimiterTablePopulateSql(l))
//		}
//	}
//
//	conn, err := m.pool.Acquire(ctx)
//	if err != nil {
//		return err
//	}
//	defer conn.Release()
//
//	_, err = db_local.ExecuteSqlWithArgsInTransaction(ctx, conn.Conn(), queries...)
//	return err
//}
