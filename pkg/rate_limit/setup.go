package ratelimit

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func GetPopulateRateLimiterSql(ctx context.Context, settings *modconfig.RateLimiter) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`INSERT INTO %s.%s (
name,
plugin,
bucket_size,
fill_rate,
scope,
"where"
)
	VALUES($1,$2,$3,$4,$5,$6)`, constants.InternalSchema, constants.ServerSettingsTable),
		Args: []any{
			settings.Name,
			settings.Plugin,
			settings.BucketSize,
			settings.FillRate,
			settings.Scope,
			settings.Where,
		},
	}
}

func CreateRateLimiterTable(ctx context.Context) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
				name TEXT NOT NULL,
				plugin TEXT NOT NULL,
				bucket_size INTEGER NOT NULL,
				fill_rate DECIMAL NOT NULL,
				scope TEXT[] NOT NULL,
				"where" TEXT NOT NULL
		);`, constants.InternalSchema, constants.RateLimiterDefinitionTable),
	}
}

func GrantsOnRateLimiterTable(ctx context.Context) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`GRANT SELECT ON TABLE %s.%s to %s;`,
			constants.InternalSchema,
			constants.RateLimiterDefinitionTable,
			constants.DatabaseUsersRole,
		),
	}
}

func DropRateLimiterTable(ctx context.Context) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`DROP TABLE IF EXISTS %s.%s;`,
			constants.InternalSchema,
			constants.RateLimiterDefinitionTable,
		),
	}
}
