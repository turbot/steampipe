package db_client

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/query/queryresult"
)

// query the database to get the current search path for the steampipe user
func (c *DbClient) GetSteampipeUserSearchPath(ctx context.Context) ([]string, error) {
	// get a database connection directly
	// NOTE: we cannot simply call ExecSync as that would call AcquireSession which calls this function
	databaseConnection, _, err := c.getDatabaseConnectionWithRetries(ctx)
	if err != nil {
		return nil, err
	}
	defer databaseConnection.Close()
	return c.GetSteampipeUserSearchPathForSession(ctx, databaseConnection)
}

// query the database to get the current search path for the steampipe user
func (c *DbClient) GetSteampipeUserSearchPathForSession(ctx context.Context, databaseConnection *sql.Conn) ([]string, error) {
	res := databaseConnection.QueryRowContext(ctx, fmt.Sprintf("select useconfig[1] from pg_user where usename='%s'", constants.DatabaseUser))
	if res.Err() != nil {
		return nil, res.Err()
	}
	pathAsString := ""
	if err := res.Scan(&pathAsString); err != nil {
		return nil, fmt.Errorf("failed to read the current user search path: %s", err.Error())
	}
	return c.buildSearchPathResult(pathAsString)
}

// GetCurrentSearchPath implements Client
// query the database to get the current session search path
func (c *DbClient) GetCurrentSearchPath(ctx context.Context) ([]string, error) {
	res, err := c.ExecuteSync(ctx, "show search_path")
	if err != nil {
		return nil, err
	}
	pathAsString, ok := res.Rows[0].(*queryresult.RowResult).Data[0].(string)
	if !ok {
		return nil, fmt.Errorf("failed to read the current search path: %s", err.Error())
	}
	return c.buildSearchPathResult(pathAsString)
}

func (c *DbClient) GetCurrentSearchPathForDbConnection(ctx context.Context, databaseConnection *sql.Conn) ([]string, error) {
	res := databaseConnection.QueryRowContext(ctx, "show search_path")
	if res.Err() != nil {
		return nil, res.Err()
	}
	pathAsString := ""
	if err := res.Scan(&pathAsString); err != nil {
		return nil, fmt.Errorf("failed to read the current search path: %s", err.Error())
	}
	return c.buildSearchPathResult(pathAsString)
}

func (c *DbClient) buildSearchPathResult(pathAsString string) ([]string, error) {
	var currentSearchPath []string

	// if this is called from GetSteampipeUserSearchPath the result will be prefixed by "search_path="
	pathAsString = strings.TrimPrefix(pathAsString, "search_path=")

	// split
	currentSearchPath = strings.Split(pathAsString, ",")

	// unescape
	for idx, p := range currentSearchPath {
		p = strings.Join(strings.Split(p, "\""), "")
		p = strings.TrimSpace(p)
		currentSearchPath[idx] = p
	}
	return currentSearchPath, nil
}

// SetRequiredSessionSearchPath implements Client
// if either a search-path or search-path-prefix is set in config, set the search path
// (otherwise fall back to user search path)
// this just sets the required search path for this client
// - when creating a database session, we will actually set the searchPath
func (c *DbClient) SetRequiredSessionSearchPath(ctx context.Context) error {
	requiredSearchPath := viper.GetStringSlice(constants.ArgSearchPath)
	searchPathPrefix := viper.GetStringSlice(constants.ArgSearchPathPrefix)

	searchPath, err := c.ContructSearchPath(ctx, requiredSearchPath, searchPathPrefix)
	if err != nil {
		return err
	}

	// store search path on the client before escaping
	c.searchPath = searchPath

	// escape the schema
	c.requiredSessionSearchPath = db_common.PgEscapeSearchPath(searchPath)

	return err
}

func (c *DbClient) ContructSearchPath(ctx context.Context, requiredSearchPath, searchPathPrefix []string) ([]string, error) {
	// if a search path was passed, add 'internal' to the end
	if len(requiredSearchPath) > 0 {
		// add 'internal' schema as last schema in the search path
		requiredSearchPath = append(requiredSearchPath, constants.FunctionSchema)
		c.customSearchPath = true
	} else {
		// so no search path was set in config - use the user search path
		steampipeUserSearchPath, err := c.GetSteampipeUserSearchPath(ctx)
		if err != nil {
			return nil, err
		}
		// cache the current user search path in the client
		c.customSearchPath = false
		requiredSearchPath = steampipeUserSearchPath
	}

	// add in the prefix if present
	requiredSearchPath = c.addSearchPathPrefix(searchPathPrefix, requiredSearchPath)

	return requiredSearchPath, nil
}

// ensure the search path for the database session is as required
func (c *DbClient) ensureSessionSearchPath(ctx context.Context, session *db_common.DatabaseSession) error {
	log.Printf("[TRACE] ensureSessionSearchPath")
	// first, if we are NOT using a custom search path, reload the user search path
	if !c.customSearchPath {
		log.Printf("[TRACE] not using a custom search path - reload the steampipe user search path")
		userSearchPath, err := c.GetSteampipeUserSearchPathForSession(ctx, session.Connection)
		if err != nil {
			return err
		}
		c.requiredSessionSearchPath = userSearchPath
		log.Printf("[TRACE] updated the required search path to %s", strings.Join(userSearchPath, ","))
	}

	// now determine whethe rthe session search path is the same as the required search path
	// if so, return
	if strings.Join(session.SearchPath, ",") == strings.Join(c.requiredSessionSearchPath, ",") {
		log.Printf("[TRACE] session search path is already correct - nothing to do")
		return nil
	}

	// so we need to set the search path
	log.Printf("[TRACE] session search path will be updated to  %s", strings.Join(c.requiredSessionSearchPath, ","))

	q := fmt.Sprintf("set search_path to %s", strings.Join(c.requiredSessionSearchPath, ","))
	_, err := session.Connection.ExecContext(ctx, q)
	if err == nil {
		// update the session search path property
		session.SearchPath = c.requiredSessionSearchPath
	}
	return err
}

func (c *DbClient) addSearchPathPrefix(searchPathPrefix []string, searchPath []string) []string {
	if len(searchPathPrefix) > 0 {
		prefixedSearchPath := searchPathPrefix
		for _, p := range searchPath {
			if !helpers.StringSliceContains(prefixedSearchPath, p) {
				prefixedSearchPath = append(prefixedSearchPath, p)
			}
		}
		searchPath = prefixedSearchPath
	}
	return searchPath
}
