package db_client

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe/pkg/constants"
)

// SetCacheTtl implements Client
func (c *DbClient) SetCacheTtl(ctx context.Context, duration time.Duration) error {
	return c.executeCacheCommand(ctx, "cache_ttl", fmt.Sprintf("%f", duration.Seconds()))
}

// CacheOn implements Client
func (c *DbClient) CacheOn(ctx context.Context) error {
	// insert into steampipe_command.settings ("name","value") values ('cache','true')
	return c.executeCacheCommand(ctx, "cache", "true")
}

// CacheOff implements Client
func (c *DbClient) CacheOff(ctx context.Context) error {
	// insert into steampipe_command.settings ("name","value") values ('cache','false')
	return c.executeCacheCommand(ctx, "cache", "false")
}

// CacheClear implements Client
func (c *DbClient) CacheClear(ctx context.Context) error {
	// insert into steampipe_command.settings ("name","value") values ('cache_clear_time','')
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
