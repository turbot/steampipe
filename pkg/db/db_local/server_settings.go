package db_local

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/serversettings"
	"github.com/turbot/steampipe/pkg/version"
)

// setupServerSettingsTable creates a new read-only table with information in the current
// settings the service has been started with.
//
// The table also includes the CLI and FDW versions for reference
func setupServerSettingsTable(ctx context.Context, conn *pgx.Conn) error {
	settings := serversettings.ServerSettings{
		SteampipeVersion: version.VersionString,
		FdwVersion:       constants.FdwVersion,
		StartTime:        time.Now().UTC(),
		CacheEnabled:     viper.GetBool(constants.ArgServiceCacheEnabled),
		CacheMaxTtl:      viper.GetInt(constants.ArgCacheMaxTtl),
		CacheMaxSizeMb:   viper.GetInt(constants.ArgMaxCacheSizeMb),
	}
	return settings.SetupTable(ctx, conn)
}
