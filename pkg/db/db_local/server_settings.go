package db_local

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/version"
)

type ServerSettingKey string

const (
	ServerSettingSteampipeVersion ServerSettingKey = "steampipe_version"
	ServerSettingFdwersion        ServerSettingKey = "fdw_version"
	ServerSettingStartTime        ServerSettingKey = "start_time"
	ServerSettingCacheEnabled     ServerSettingKey = "cache_enabled"
	ServerSettingCacheMaxTtl      ServerSettingKey = "cache_max_ttl"
	ServerSettingCacheMaxSizeMb   ServerSettingKey = "cache_max_size_mb"
)

// initializeServerSettingsTable creates a new read-only table with information in the current
// settings the service has been started with.
//
// The table also includes the CLI and FDW versions for reference
func initializeServerSettingsTable(ctx context.Context, conn *pgx.Conn) error {
	utils.LogTime("db_local.initializeServerSettingsTable start")
	defer utils.LogTime("db_local.initializeServerSettingsTable end")

	settings := map[ServerSettingKey]any{
		ServerSettingSteampipeVersion: version.VersionString,
		ServerSettingFdwersion:        constants.FdwVersion,
		ServerSettingStartTime:        time.Now().UTC().Format(time.RFC3339),
		ServerSettingCacheEnabled:     viper.GetBool(constants.ArgServiceCacheEnabled),
		ServerSettingCacheMaxTtl:      viper.GetInt(constants.ArgCacheMaxTtl),
		ServerSettingCacheMaxSizeMb:   viper.GetInt(constants.ArgMaxCacheSizeMb),
	}

	// start with a clean slate
	queries := []db_common.QueryWithArgs{
		// drop the old table (alternative is "if exists then truncate" which is more expensive)
		// this also allows us to modify the table structure without having to go through complex
		// migrations
		getServerSettingsTableDropSQL(ctx),
		// create a new one
		getServerSettingsTableCreateSQL(ctx),
		// grants
		getServerSettingsTableGrantSQL(ctx),
	}

	queries = append(queries, getServerSettingsRowSql(ctx, settings)...)

	_, err := ExecuteSqlWithArgsInTransaction(ctx, conn, queries...)
	return err
}

func getServerSettingsRowSql(_ context.Context, settings map[ServerSettingKey]any) []db_common.QueryWithArgs {
	queries := []db_common.QueryWithArgs{}
	for name, value := range settings {
		dataType := "text"
		kind := reflect.TypeOf(value).Kind()
		switch kind {
		case reflect.Bool:
			dataType = "bool"
		case reflect.String:
			dataType = "text"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			dataType = "integer"
		default:
			panic(fmt.Sprintf("unknown type: %s", kind.String()))
		}

		queries = append(queries, db_common.QueryWithArgs{
			Query: fmt.Sprintf(
				`INSERT INTO %s.%s (name,value) VALUES ($1,TO_JSONB($2::%s))`,
				constants.InternalSchema,
				constants.ServerSettingsTable,
				dataType,
			),
			Args: []any{name, value},
		})

	}
	return queries
}

func getServerSettingsTableGrantSQL(_ context.Context) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`GRANT SELECT ON TABLE %s.%s to %s;`,
			constants.InternalSchema,
			constants.ServerSettingsTable,
			constants.DatabaseUsersRole,
		),
	}
}

func getServerSettingsTableCreateSQL(_ context.Context) db_common.QueryWithArgs {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		name TEXT PRIMARY KEY,
		value JSONB NOT NULL
		);`, constants.InternalSchema, constants.ServerSettingsTable)

	return db_common.QueryWithArgs{Query: query}
}

func getServerSettingsTableDropSQL(_ context.Context) db_common.QueryWithArgs {
	query := fmt.Sprintf(
		`DROP TABLE IF EXISTS %s.%s;`,
		constants.InternalSchema,
		constants.ServerSettingsTable,
	)

	return db_common.QueryWithArgs{Query: query}
}
