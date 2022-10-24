package db_client

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/query/queryresult"
)

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

// GetCurrentSearchPathForDbConnection queries the database to get the current session search path
// using the given connection
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

	// escape the search path
	c.requiredSessionSearchPath = db_common.PgEscapeSearchPath(searchPath)

	return err
}

// GetRequiredSessionSearchPath implements Client
func (c *DbClient) GetRequiredSessionSearchPath() []string {
	return c.requiredSessionSearchPath
}

func (c *DbClient) ContructSearchPath(ctx context.Context, customSearchPath, searchPathPrefix []string) ([]string, error) {
	// strip empty elements from search path and prefix
	customSearchPath = helpers.RemoveFromStringSlice(customSearchPath, "")
	searchPathPrefix = helpers.RemoveFromStringSlice(searchPathPrefix, "")

	// store custom search path and search path prefix
	c.searchPathPrefix = searchPathPrefix
	var requiredSearchPath []string
	// if a search path was passed, add 'internal' to the end
	if len(customSearchPath) > 0 {
		// add 'internal' schema as last schema in the search path
		customSearchPath = append(customSearchPath, constants.FunctionSchema)
		// store the modified custom search path on the client
		c.customSearchPath = customSearchPath
		requiredSearchPath = c.customSearchPath
	} else {
		// so no search path was set in config
		c.customSearchPath = nil
		// use the default search path
		requiredSearchPath = c.GetDefaultSearchPath(ctx)
	}

	// add in the prefix if present
	requiredSearchPath = c.addSearchPathPrefix(searchPathPrefix, requiredSearchPath)

	return requiredSearchPath, nil
}

// ensure the search path for the database session is as required
func (c *DbClient) ensureSessionSearchPath(ctx context.Context, session *db_common.DatabaseSession) error {
	log.Printf("[TRACE] ensureSessionSearchPath")
	// first, if we are NOT using a custom search path, reload the steampipe user search path
	if len(c.customSearchPath) == 0 {
		// rebuild required search path using the prefix, if any
		requiredSearchPath := c.addSearchPathPrefix(c.searchPathPrefix, c.GetDefaultSearchPath(ctx))
		// escape the required search path and store on client
		c.requiredSessionSearchPath = db_common.PgEscapeSearchPath(requiredSearchPath)
		log.Printf("[TRACE] updated the required search path to %s", strings.Join(c.requiredSessionSearchPath, ","))
	}

	// now determine whether the session search path is the same as the required search path
	// if so, return
	if strings.Join(session.SearchPath, ",") == strings.Join(c.requiredSessionSearchPath, ",") {
		log.Printf("[TRACE] session search path is already correct - nothing to do")
		return nil
	}

	// so we need to set the search path
	log.Printf("[TRACE] session search path will be updated to  %s", strings.Join(c.requiredSessionSearchPath, ","))

	q := fmt.Sprintf("set search_path to %s", strings.Join(c.requiredSessionSearchPath, ","))
	_, err := session.Connection.Exec(ctx, q)
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
