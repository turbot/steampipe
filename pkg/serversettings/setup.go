package serversettings

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

func GetPopulateServerSettingsSql(ctx context.Context, settings db_common.ServerSettings) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`INSERT INTO %s.%s (
start_time,
steampipe_version,
fdw_version,
cache_max_ttl,
cache_max_size_mb,
cache_enabled)
	VALUES($1,$2,$3,$4,$5,$6)`, constants.InternalSchema, constants.ServerSettingsTable),
		Args: []any{
			settings.StartTime,
			settings.SteampipeVersion,
			settings.FdwVersion,
			settings.CacheMaxTtl,
			settings.CacheMaxSizeMb,
			settings.CacheEnabled,
		},
	}
}

func CreateServerSettingsTable(ctx context.Context) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
start_time TIMESTAMPTZ NOT NULL,
steampipe_version TEXT NOT NULL,
fdw_version TEXT NOT NULL,
cache_max_ttl INTEGER NOT NULL,
cache_max_size_mb INTEGER NOT NULL,
cache_enabled BOOLEAN NOT NULL
		);`, constants.InternalSchema, constants.ServerSettingsTable),
	}
}

func GrantsOnServerSettingsTable(ctx context.Context) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`GRANT SELECT ON TABLE %s.%s to %s;`,
			constants.InternalSchema,
			constants.ServerSettingsTable,
			constants.DatabaseUsersRole,
		),
	}
}

func DropServerSettingsTable(ctx context.Context) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`DROP TABLE IF EXISTS %s.%s;`,
			constants.InternalSchema,
			constants.ServerSettingsTable,
		),
	}
}
