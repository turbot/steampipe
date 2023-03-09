package db_client

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"
)

// CacheOn implements Client
func (c *DbClient) CacheOn(ctx context.Context) error {
	return c.executeCacheCommand(ctx, constants.CommandCacheOn)
}

// CacheOff implements Client
func (c *DbClient) CacheOff(ctx context.Context) error {
	return c.executeCacheCommand(ctx, constants.CommandCacheOff)
}

// CacheClear implements Client
func (c *DbClient) CacheClear(ctx context.Context) error {
	return c.executeCacheCommand(ctx, constants.CommandCacheClear)
}

func (c *DbClient) executeCacheCommand(ctx context.Context, controlCommand string) error {
	_, err := c.pool.Exec(ctx, fmt.Sprintf(
		"insert into %s.%s (%s) values ('%s')",
		constants.CommandSchema,
		constants.CommandTableCache,
		constants.CommandTableCacheOperationColumn,
		controlCommand,
	))
	return err
}
