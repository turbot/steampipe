package introspection

import (
	"fmt"

	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

func GetRateLimiterTableCreateSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
				name TEXT,
				plugin TEXT,
				plugin_instance TEXT NULL,
				source_type TEXT,
				status TEXT,
				bucket_size INTEGER,
				fill_rate REAL ,
				max_concurrency INTEGER,
				scope JSONB,
				"where" TEXT,
				file_name TEXT, 
				start_line_number INTEGER, 
				end_line_number INTEGER 
		);`, constants.InternalSchema, constants.RateLimiterDefinitionTable),
	}
}

func GetRateLimiterTableDropSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`DROP TABLE IF EXISTS %s.%s;`,
			constants.InternalSchema,
			constants.RateLimiterDefinitionTable,
		),
	}
}

func GetRateLimiterTablePopulateSql(settings *plugin.RateLimiter) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`INSERT INTO %s.%s (
"name",
plugin,
plugin_instance,
source_type,
status,
bucket_size,
fill_rate,
max_concurrency,
scope,
"where",
file_name,
start_line_number,
end_line_number
)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, constants.InternalSchema, constants.RateLimiterDefinitionTable),
		Args: []any{
			settings.Name,
			settings.Plugin,
			settings.PluginInstance,
			settings.Source,
			settings.Status,
			settings.BucketSize,
			settings.FillRate,
			settings.MaxConcurrency,
			settings.Scope,
			settings.Where,
			settings.FileName,
			settings.StartLineNumber,
			settings.EndLineNumber,
		},
	}
}

func GetRateLimiterTableGrantSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`GRANT SELECT ON TABLE %s.%s to %s;`,
			constants.InternalSchema,
			constants.RateLimiterDefinitionTable,
			constants.DatabaseUsersRole,
		),
	}
}
