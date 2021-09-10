package db_client

import (
	"context"
	"fmt"
	"strings"

	"github.com/turbot/steampipe/db/db_common"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
)

// GetCurrentSearchPath implements Client
// query the database to get the current search path
func (c *DbClient) GetCurrentSearchPath() ([]string, error) {
	var currentSearchPath []string
	var pathAsString string
	row := c.dbClient.QueryRow("show search_path")
	if row.Err() != nil {
		return nil, row.Err()
	}
	err := row.Scan(&pathAsString)
	if err != nil {
		return nil, err
	}
	currentSearchPath = strings.Split(pathAsString, ",")
	// unescape search path
	for idx, p := range currentSearchPath {
		p = strings.Join(strings.Split(p, "\""), "")
		p = strings.TrimSpace(p)
		currentSearchPath[idx] = p
	}
	return currentSearchPath, nil
}

// SetSessionSearchPath implements Client
// sets the search path for this client
// if either a search-path or search-path-prefix is set in config, set the search path
// (otherwise fall back to service search path)
func (c *DbClient) SetSessionSearchPath(currentSearchPath ...string) error {
	searchPath := viper.GetStringSlice(constants.ArgSearchPath)
	searchPathPrefix := viper.GetStringSlice(constants.ArgSearchPathPrefix)

	// if a search path was passed, add 'internal' to the end
	if len(searchPath) > 0 {
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.FunctionSchema)
	} else {
		// so no search path was set in config - use the current search poath

		// if this function is called from local db client, it will pass in the current search path
		// we must do this as the local client will reload the service search path
		if len(currentSearchPath) == 0 {
			// no current search path was passed in - fetch it
			var err error
			if currentSearchPath, err = c.GetCurrentSearchPath(); err != nil {
				return err
			}
		}
		searchPath = currentSearchPath
	}

	// add in the prefix if present
	searchPath = c.addSearchPathPrefix(searchPathPrefix, searchPath)

	// store search path on the client before escaping
	c.schemaMetadata.SearchPath = searchPath

	// escape the schema
	searchPath = db_common.PgEscapeSearchPath(searchPath)

	// now construct and execute the query
	q := fmt.Sprintf("set search_path to %s", strings.Join(searchPath, ","))
	_, err := c.ExecuteSync(context.Background(), q, true)
	if err != nil {
		return err
	}
	return nil
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
