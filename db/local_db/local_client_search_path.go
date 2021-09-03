package local_db

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/turbot/steampipe/db/db_common"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
)

// GetCurrentSearchPath implements DbClient
// query the database to get the current search path
func (c *LocalClient) GetCurrentSearchPath() ([]string, error) {
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

// SetClientSearchPath implements DbClient
//sets the search path for this client
// if either a search-path or search-path-prefix is set in config, set the search path
// (otherwise fall back to service search path)
func (c *LocalClient) SetClientSearchPath() error {
	searchPath := viper.GetStringSlice(constants.ArgSearchPath)
	searchPathPrefix := viper.GetStringSlice(constants.ArgSearchPathPrefix)

	// if a search path was passed, add 'internal' to the end
	if len(searchPath) > 0 {
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.FunctionSchema)
	} else {
		// so no search path was set in config
		// in this case we need to load the existing service search path
		var err error
		if searchPath, err = getCurrentSearchPath(); err != nil {
			return err
		}
	}

	// add in the prefix if present
	searchPath = c.addSearchPathPrefix(searchPathPrefix, searchPath)

	// store search path on the client before escaping
	c.schemaMetadata.SearchPath = searchPath

	// escape the schema
	searchPath = escapeSearchPath(searchPath)

	// now construct and execute the query
	q := fmt.Sprintf("set search_path to %s", strings.Join(searchPath, ","))
	_, err := c.ExecuteSync(context.Background(), q, true)
	if err != nil {
		return err
	}
	return nil
}

func getCurrentSearchPath() ([]string, error) {
	// NOTE: create a new client to do this so we respond to any recent changes in service search path
	// (as the service search path may have changed  after creating client 'c', e.g. if connections have changed)
	c, err := NewLocalClient(constants.InvokerService)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return c.GetCurrentSearchPath()
}

// SetServiceSearchPath sets the search path for the db service (by setting it on the steampipe user)
func (c *LocalClient) SetServiceSearchPath() error {
	var searchPath []string

	// is there a service search path in the config?
	// check ConfigKeyDatabaseSearchPath config (this is the value specified in the database config)
	if viper.IsSet(constants.ConfigKeyDatabaseSearchPath) {
		searchPath = viper.GetStringSlice(constants.ConfigKeyDatabaseSearchPath)
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.FunctionSchema)
	} else {
		// no config set - set service search path to default
		searchPath = c.getDefaultSearchPath()
	}

	// escape the schema names
	searchPath = escapeSearchPath(searchPath)

	log.Println("[TRACE] setting service search path to", searchPath)

	// now construct and execute the query
	query := fmt.Sprintf(
		"alter user %s set search_path to %s;",
		constants.DatabaseUser,
		strings.Join(searchPath, ","),
	)
	_, err := c.ExecuteSync(context.Background(), query, true)
	return err
}

func (c *LocalClient) addSearchPathPrefix(searchPathPrefix []string, searchPath []string) []string {
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

// build default search path from the connection schemas, bookended with public and internal
func (c *LocalClient) getDefaultSearchPath() []string {
	searchPath := c.schemaMetadata.GetSchemas()
	sort.Strings(searchPath)
	// add the 'public' schema as the first schema in the search_path. This makes it
	// easier for users to build and work with their own tables, and since it's normally
	// empty, doesn't make using steampipe tables any more difficult.
	searchPath = append([]string{"public"}, searchPath...)
	// add 'internal' schema as last schema in the search path
	searchPath = append(searchPath, constants.FunctionSchema)

	return searchPath
}

// apply postgres escaping to search path and remove whitespace
func escapeSearchPath(searchPath []string) []string {
	res := make([]string, len(searchPath))
	for idx, path := range searchPath {
		res[idx] = db_common.PgEscapeName(strings.TrimSpace(path))
	}
	return res
}
