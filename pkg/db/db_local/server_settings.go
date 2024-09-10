package db_local

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/serversettings"
	"github.com/turbot/steampipe/pkg/version"
)

// setupServerSettingsTable creates a new read-only table with information in the current
// settings the service has been started with.
//
// The table also includes the CLI and FDW versions for reference
func setupServerSettingsTable(ctx context.Context, conn *pgx.Conn) error {
	settings := db_common.ServerSettings{
		StartTime:        time.Now(),
		SteampipeVersion: version.VersionString,
		FdwVersion:       constants.FdwVersion,
		CacheMaxTtl:      viper.GetInt(pconstants.ArgCacheMaxTtl),
		CacheMaxSizeMb:   viper.GetInt(pconstants.ArgMaxCacheSizeMb),
		CacheEnabled:     viper.GetBool(pconstants.ArgServiceCacheEnabled),
	}

	queries := []db_common.QueryWithArgs{
		serversettings.DropServerSettingsTable(ctx),
		serversettings.CreateServerSettingsTable(ctx),
		serversettings.GrantsOnServerSettingsTable(ctx),
		serversettings.GetPopulateServerSettingsSql(ctx, settings),
	}

	log.Println("[TRACE] saved server settings:", settings)

	_, err := ExecuteSqlWithArgsInTransaction(ctx, conn, queries...)
	return err
}
