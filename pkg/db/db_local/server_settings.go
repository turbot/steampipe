package db_local

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/version"
)

// initializeServerSettingsTable creates a new read-only table with information in the current
// settings the service has been started with.
//
// The table also includes the CLI and FDW versions for reference
func initializeServerSettingsTable(ctx context.Context, conn *pgx.Conn) error {
	settings := db_common.ServerSettings{
		SteampipeVersion: version.VersionString,
		FdwVersion:       constants.FdwVersion,
		StartTime:        time.Now(),
		CacheEnabled:     viper.GetBool(constants.ArgServiceCacheEnabled),
		CacheMaxTtl:      viper.GetInt(constants.ArgCacheMaxTtl),
		CacheMaxSizeMb:   viper.GetInt(constants.ArgMaxCacheSizeMb),
	}
	queries := settings.SetupSql(ctx)
	_, err := ExecuteSqlWithArgsInTransaction(ctx, conn, queries...)
	return err
}
