package pluginmanager_service

import (
	"context"
	"log"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	ratelimit "github.com/turbot/steampipe/pkg/rate_limit"
)

func (m *PluginManager) refreshRateLimiterTable(ctx context.Context) error {
	log.Println("[TRACE] >>> refreshRateLimiterTable")
	defer log.Println("[TRACE] <<< refreshRateLimiterTable")

	queries := []db_common.QueryWithArgs{
		ratelimit.DropRateLimiterTable(ctx),
		ratelimit.CreateRateLimiterTable(ctx),
		ratelimit.GrantsOnRateLimiterTable(ctx),
	}

	for _, limiter := range m.limiters {
		queries = append(queries, ratelimit.GetPopulateRateLimiterSql(ctx, limiter))
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
