package local_db

import (
	"fmt"

	"github.com/turbot/steampipe/constants"
)

func (c *LocalClient) executeCacheCommand(controlCommand string) error {
	_, err := c.dbClient.Exec(fmt.Sprintf(
		"insert into %s.%s (%s) values ('%s')",
		constants.CommandSchema,
		constants.CacheCommandTable,
		constants.CacheCommandOperationColumn,
		controlCommand,
	))
	return err
}

func (c *LocalClient) CacheOn() error {
	return c.executeCacheCommand(constants.CommandCacheOn)
}

func (c *LocalClient) CacheOff() error {
	return c.executeCacheCommand(constants.CommandCacheOff)
}

func (c *LocalClient) CacheClear() error {
	return c.executeCacheCommand(constants.CommandCacheClear)
}
