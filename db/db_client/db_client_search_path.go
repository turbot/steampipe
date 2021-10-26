package db_client

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/query/queryresult"
)

// GetCurrentSearchPath implements Client
// query the database to get the current search path
func (c *DbClient) GetCurrentSearchPath() ([]string, error) {
	var currentSearchPath []string
	var pathAsString string
	rows, err := c.ExecuteSync(context.Background(), "show search_path", false)
	if err != nil {
		return nil, err
	}
	pathAsString, ok := rows.Rows[0].(*queryresult.RowResult).Data[0].(string)
	if !ok {
		return nil, fmt.Errorf("error during extracting the search path from service")
	}
	fmt.Println(pathAsString)
	currentSearchPath = strings.Split(pathAsString, ",")
	fmt.Println(currentSearchPath)
	// unescape search path
	for idx, p := range currentSearchPath {
		p = strings.Join(strings.Split(p, "\""), "")
		p = strings.TrimSpace(p)
		currentSearchPath[idx] = p
	}
	fmt.Println(currentSearchPath)
	return currentSearchPath, nil
}

// SetSessionSearchPath implements Client
// sets the search path for this client
// if either a search-path or search-path-prefix is set in config, set the search path
// (otherwise fall back to user search path)
func (c *DbClient) SetSessionSearchPath(currentSearchPath ...string) error {
	requiredSearchPath := viper.GetStringSlice(constants.ArgSearchPath)
	searchPathPrefix := viper.GetStringSlice(constants.ArgSearchPathPrefix)

	searchPath, err := c.ContructSearchPath(requiredSearchPath, searchPathPrefix, currentSearchPath)
	if err != nil {
		return err
	}

	// store search path on the client before escaping
	c.schemaMetadata.SearchPath = searchPath

	// escape the schema
	c.requiredSessionSearchPath = db_common.PgEscapeSearchPath(searchPath)

	return err
}

func (c *DbClient) ContructSearchPath(requiredSearchPath []string, searchPathPrefix []string, currentSearchPath []string) ([]string, error) {
	// if a search path was passed, add 'internal' to the end
	if len(requiredSearchPath) > 0 {
		// add 'internal' schema as last schema in the search path
		requiredSearchPath = append(requiredSearchPath, constants.FunctionSchema)
	} else {
		// so no search path was set in config - use the current search poath

		// if this function is called from local db client, it will pass in the current search path
		// we must do this as the local client will reload the user search path
		if len(currentSearchPath) == 0 {
			// no current search path was passed in - fetch it
			var err error
			if currentSearchPath, err = c.GetCurrentSearchPath(); err != nil {
				return nil, err
			}
		}
		requiredSearchPath = currentSearchPath
	}

	// add in the prefix if present
	requiredSearchPath = c.addSearchPathPrefix(searchPathPrefix, requiredSearchPath)

	return requiredSearchPath, nil
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
