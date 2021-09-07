package local_db

import (
	"fmt"

	"github.com/turbot/steampipe/constants"
)

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

func (c *DbClient) CacheOn() error {
	return c.executeCacheCommand(constants.CommandCacheOn)
}

func (c *DbClient) CacheOff() error {
	return c.executeCacheCommand(constants.CommandCacheOff)
}

func (c *DbClient) CacheClear() error {
	return c.executeCacheCommand(constants.CommandCacheClear)
}
