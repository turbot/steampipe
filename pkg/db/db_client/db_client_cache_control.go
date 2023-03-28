package db_client

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/turbot/steampipe/pkg/constants"
)

// SetCacheTtl implements Client
func (c *DbClient) SetCacheTtl(ctx context.Context, duration time.Duration) error {
	duration = duration.Truncate(time.Second)
	// we need to use strconv here since a simple fmt.Sprintf with %f results in trailing zeros
	// which is a problem when we parse this out as an int in the FDW
	// the '-1' in the FormatFloat makes sure that there are only as many trailing
	// zeros as is required to accurately stringify this float (which is none because we truncate)
	return c.executeCacheCommand(ctx, "cache_ttl", strconv.FormatFloat(duration.Seconds(), 'f', -1, 64))
}

// CacheOn implements Client
func (c *DbClient) CacheOn(ctx context.Context) error {
	return c.executeCacheCommand(ctx, "cache", "true")
}

// CacheOff implements Client
func (c *DbClient) CacheOff(ctx context.Context) error {
	return c.executeCacheCommand(ctx, "cache", "false")
}

// CacheClear implements Client
func (c *DbClient) CacheClear(ctx context.Context) error {
	return c.executeCacheCommand(ctx, "cache_clear_time", "")
}

func (c *DbClient) executeCacheCommand(ctx context.Context, settingName string, settingValue string) error {
	_, err := c.pool.Exec(ctx, fmt.Sprintf(
		"insert into %s.%s (%s,%s) values ('%s','%s')",
		constants.CommandSchema,
		constants.CommandTableSettings,
		constants.CommandTableSettingsKeyColumn,
		constants.CommandTableSettingsValueColumn,
		settingName,
		settingValue,
	))
	return err
}
