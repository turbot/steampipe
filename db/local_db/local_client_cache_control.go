package local_db

import (
	"fmt"

	"github.com/turbot/steampipe/constants"
)

func (c *LocalClient) executeCacheControlStatementWith(controlCommand string) error {
	_, err := c.dbClient.Exec(fmt.Sprintf(
		"select %s from %s.%s",
		controlCommand,
		constants.FDWCommandSchema,
		constants.FDWCommandTable,
	))
	return err
}

func (c *LocalClient) CacheOn() error {
	return c.executeCacheControlStatementWith(constants.FDWCacheOnCommand)
}

func (c *LocalClient) CacheOff() error {
	return c.executeCacheControlStatementWith(constants.FDWCacheOffCommand)
}

func (c *LocalClient) CacheClear() error {
	return c.executeCacheControlStatementWith(constants.FDWCacheClearCommand)
}
