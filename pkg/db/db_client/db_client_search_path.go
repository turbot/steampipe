package db_client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

// SetRequiredSessionSearchPath implements Client
// if either a search-path or search-path-prefix is set in config, set the search path
// (otherwise fall back to user search path)
// this just sets the required search path for this client
// - when creating a database session, we will actually set the searchPath
func (c *DbClient) SetRequiredSessionSearchPath(ctx context.Context) error {
	configuredSearchPath := viper.GetStringSlice(constants.ArgSearchPath)
	searchPathPrefix := viper.GetStringSlice(constants.ArgSearchPathPrefix)

	// strip empty elements from search path and prefix
	configuredSearchPath = helpers.RemoveFromStringSlice(configuredSearchPath, "")
	searchPathPrefix = helpers.RemoveFromStringSlice(searchPathPrefix, "")

	// default required path to user search path
	requiredSearchPath := c.userSearchPath

	// store custom search path and search path prefix
	c.searchPathPrefix = searchPathPrefix

	// if a search path was passed, add 'internal' to the end
	if len(configuredSearchPath) > 0 {
		// add 'internal' schema as last schema in the search path
		requiredSearchPath = append(configuredSearchPath, constants.InternalSchema)
	}

	// add in the prefix if present
	requiredSearchPath = db_common.AddSearchPathPrefix(searchPathPrefix, requiredSearchPath)

	// if either configuredSearchPath or searchPathPrefix are set, store requiredSearchPath as customSearchPath
	if len(configuredSearchPath)+len(searchPathPrefix) > 0 {
		c.customSearchPath = requiredSearchPath
	} else {
		// otherwise clear it
		c.customSearchPath = nil
	}

	return nil
}

func (c *DbClient) LoadUserSearchPath(ctx context.Context) error {
	conn, _, err := c.GetDatabaseConnectionWithRetries(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	return c.loadUserSearchPath(ctx, conn.Conn())
}

func (c *DbClient) loadUserSearchPath(ctx context.Context, connection *pgx.Conn) error {
	// load the user search path
	userSearchPath, err := db_common.GetUserSearchPath(ctx, connection)
	if err != nil {
		return err
	}
	// update the cached value
	c.userSearchPath = userSearchPath
	return nil
}

// GetRequiredSessionSearchPath implements Client
func (c *DbClient) GetRequiredSessionSearchPath() []string {
	if c.customSearchPath != nil {
		return c.customSearchPath
	}

	return c.userSearchPath
}

func (c *DbClient) GetCustomSearchPath() []string {
	return c.customSearchPath
}

// reload Steampipe config, update viper and re-set required search path
func (c *DbClient) updateRequiredSearchPath(ctx context.Context) error {
	config, errorsAndWarnings := steampipeconfig.LoadSteampipeConfig(viper.GetString(constants.ArgModLocation), "dashboard")
	if errorsAndWarnings.GetError() != nil {
		return errorsAndWarnings.GetError()
	}
	// todo review this usage of GlobalConfig
	// https://github.com/turbot/steampipe/issues/3387
	steampipeconfig.GlobalConfig = config
	cmdconfig.SetDefaultsFromConfig(steampipeconfig.GlobalConfig.ConfigMap())
	return c.SetRequiredSessionSearchPath(ctx)
}

// ensure the search path for the database session is as required
func (c *DbClient) ensureSessionSearchPath(ctx context.Context, session *db_common.DatabaseSession) error {
	log.Printf("[TRACE] ensureSessionSearchPath")

	// update the stored value of user search path
	if err := c.loadUserSearchPath(ctx, session.Connection.Conn()); err != nil {
		return err
	}

	requiredSearchPath := c.GetRequiredSessionSearchPath()

	// now determine whether the session search path is the same as the required search path
	// if so, return
	if strings.Join(session.SearchPath, ",") == strings.Join(requiredSearchPath, ",") {
		log.Printf("[TRACE] session search path is already correct - nothing to do")
		return nil
	}

	// so we need to set the search path
	log.Printf("[TRACE] session search path will be updated to  %s", strings.Join(c.customSearchPath, ","))

	_, err := session.Connection.Exec(ctx, fmt.Sprintf("set search_path to %s", strings.Join(db_common.PgEscapeSearchPath(requiredSearchPath), ",")))
	if err == nil {
		// update the session search path property
		session.SearchPath = requiredSearchPath
	}
	return err
}
