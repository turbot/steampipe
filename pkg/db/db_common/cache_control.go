package db_common

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// SetCacheTtl set the cache ttl on the client
func SetCacheTtl(ctx context.Context, duration time.Duration, connection *pgx.Conn) error {
	duration = duration.Truncate(time.Second)
	seconds := fmt.Sprint(duration.Seconds())
	return executeCacheTtlSetFunction(ctx, seconds, connection)
}

// CacheClear resets the max time on the cache
// anything below this is not accepted
func CacheClear(ctx context.Context, connection *pgx.Conn) error {
	return executeCacheSetFunction(ctx, "clear", connection)
}

// SetCacheEnabled enables/disables the cache
func SetCacheEnabled(ctx context.Context, enabled bool, connection *pgx.Conn) error {
	value := "off"
	if enabled {
		value = "on"
	}
	return executeCacheSetFunction(ctx, value, connection)
}

func executeCacheSetFunction(ctx context.Context, settingValue string, connection *pgx.Conn) error {
	return ExecuteSystemClientCall(ctx, connection, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, fmt.Sprintf(
			"select %s.%s('%s')",
			constants.InternalSchema,
			constants.FunctionCacheSet,
			settingValue,
		))
		return err
	})
}

func executeCacheTtlSetFunction(ctx context.Context, seconds string, connection *pgx.Conn) error {
	return ExecuteSystemClientCall(ctx, connection, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, fmt.Sprintf(
			"select %s.%s('%s')",
			constants.InternalSchema,
			constants.FunctionCacheSetTtl,
			seconds,
		))
		return err
	})
}
