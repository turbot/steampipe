package db_client

import (
	"fmt"

	"github.com/turbot/steampipe/constants"
)

// CacheOn implements Client
func (c *DbClient) CacheOn() error {
	return c.executeCacheCommand(constants.CommandCacheOn)
}

// CacheOff implements Client
func (c *DbClient) CacheOff() error {
	return c.executeCacheCommand(constants.CommandCacheOff)
}

// CacheClear implements Client
func (c *DbClient) CacheClear() error {
	return c.executeCacheCommand(constants.CommandCacheClear)
}

func (c *DbClient) executeCacheCommand(controlCommand string) error {
	_, err := c.dbClient.Exec(fmt.Sprintf(
		"insert into %s.%s (%s) values ('%s')",
		constants.CommandSchema,
		constants.CacheCommandTable,
		constants.CacheCommandOperationColumn,
		controlCommand,
	))
	return err
}
