package db_client

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe/pkg/constants"
)

// SetCacheTtl implements Client
func (c *DbClient) SetCacheTtl(ctx context.Context, duration time.Duration) error {
	duration = duration.Truncate(time.Second)
	seconds := int(duration.Seconds())
	return c.executeCacheCommand(ctx, constants.CommandTableSettingsCacheTtlKey, fmt.Sprint(seconds))
}

// CacheOn implements Client
func (c *DbClient) CacheOn(ctx context.Context) error {
	return c.executeCacheCommand(ctx, constants.CommandTableSettingsCacheKey, "true")
}

// CacheOff implements Client
func (c *DbClient) CacheOff(ctx context.Context) error {
	return c.executeCacheCommand(ctx, constants.CommandTableSettingsCacheKey, "false")
}

// CacheClear implements Client
func (c *DbClient) CacheClear(ctx context.Context) error {
	return c.executeCacheCommand(ctx, constants.CommandTableSettingsCacheClearTimeKey, "")
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
