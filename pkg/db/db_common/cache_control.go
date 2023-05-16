package db_common

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/constants"
)

// SetCacheTtl set the cache ttl on the client
func SetCacheTtl(ctx context.Context, duration time.Duration, connection *pgx.Conn) error {
	duration = duration.Truncate(time.Second)
	seconds := int(duration.Seconds())
	return executeCacheCommand(ctx, constants.CommandTableSettingsCacheTtlKey, fmt.Sprint(seconds), connection)
}

// CacheClear resets the max time on the cache
// anything below this is not accepted
func CacheClear(ctx context.Context, connection *pgx.Conn) error {
	return executeCacheCommand(ctx, constants.CommandTableSettingsCacheClearTimeKey, "", connection)
}

// SetCacheEnabled enables/disables the cache
func SetCacheEnabled(ctx context.Context, enabled bool, connection *pgx.Conn) error {
	return executeCacheCommand(ctx, constants.CommandTableSettingsCacheKey, fmt.Sprint(enabled), connection)
}

func executeCacheCommand(ctx context.Context, settingName string, settingValue string, connection *pgx.Conn) error {
	_, err := connection.Exec(ctx, fmt.Sprintf(
		"insert into %s.%s (%s,%s) values ('%s','%s')",
		constants.InternalSchema,
		constants.CommandTableSettings,
		constants.CommandTableSettingsKeyColumn,
		constants.CommandTableSettingsValueColumn,
		settingName,
		settingValue,
	))
	return err
}
