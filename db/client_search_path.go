package db

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
)

// set the search path for this client
// if either a search-path or search-path-prefix is set in config, set the search path
func (c *Client) setClientSearchPath() error {
	searchPath := viper.GetStringSlice(constants.ArgSearchPath)
	searchPathPrefix := viper.GetStringSlice(constants.ArgSearchPathPrefix)

	// HACK reopen db client so we take into account recent changes to service search path
	if err := c.refreshDbClient(); err != nil {
		return err
	}
	if len(searchPath) == 0 && len(searchPathPrefix) == 0 {
		return nil
	}

	// if neither search-path or search-path-prefix are set in config,
	// we _could_ just fall back to using the service search path
	// however changes to the service searhc path will not be reflected until we create a new DB client
	// also - we need to update the search path stored on the client
	// so instead, always explicitly set the client search path

	// if a search path was passed, add 'internal' to the end
	if len(searchPath) > 0 {
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.FunctionSchema)
	} else {
		// so no search path was set in config
		// in this case we need to either load the existing service search path, or build the default search path from connections

		// if the service search path is the (previous) default, then set our search path to the (new) default

		// if the current search path is the same as the previous default,
		// then we assume no search path has been explicitly set on the service
		// - in that case, just update the client search path to the NEW default
		// (this handles the case where there are new connections)
		// TODO getCurrentSearchPath does not reflect changes to service search path since the client was created
		// if this worked, we would not need to check prevDefaultSearchPath here

		searchPath, _ = c.getCurrentSearchPath()
		//// TODO add searchPathEqual function
		//if prevDefaultSearchPath != nil && reflect.DeepEqual(searchPath, prevDefaultSearchPath) {
		//	searchPath = c.getDefaultSearchPath()
		//}
	}

	// add in the prefix if present
	searchPath = c.addSearchPathPrefix(searchPathPrefix, searchPath)

	// escape the schema
	searchPath = escapeSearchPath(searchPath)

	// now construct and execute the query
	q := fmt.Sprintf("set search_path to %s", strings.Join(searchPath, ","))
	_, err := c.ExecuteSync(q)
	if err != nil {
		return err
	}

	// store search path on the client
	c.schemaMetadata.SearchPath = searchPath
	return nil
}

// set the search path for the db service (by setting it on the steampipe user)
// DO NOT set the search path the default if the existing search path is not the same as the previous default
// (as this indicates the service search path has been set either via config,
// or on the command line (which we cannot detect as it would have been in a different steampipe session)
func (c *Client) setServiceSearchPath(prevDefaultSearchPath []string) error {
	var searchPath []string

	// is there a service search path in the config?
	// check ConfigKeyDatabaseSearchPath config (this is the value specified in the database config)
	if viper.IsSet(constants.ConfigKeyDatabaseSearchPath) {
		searchPath = viper.GetStringSlice(constants.ConfigKeyDatabaseSearchPath)
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.FunctionSchema)
	} else {
		// no config set - set service search path to default

		// if the current service search path is NOT the previous default search path,
		// it means is has been explicitly set via a command line arg so we DO NOT want to update it
		searchPath, _ = c.getCurrentSearchPath()
		if prevDefaultSearchPath != nil && !reflect.DeepEqual(searchPath, prevDefaultSearchPath) {
			return nil
		}

		// so current service search path IS the same as the previous default
		// update it to the new default
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
	_, err := c.ExecuteSync(query)
	return err
}

func (c *Client) addSearchPathPrefix(searchPathPrefix []string, searchPath []string) []string {
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
func (c *Client) getDefaultSearchPath() []string {
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

// query the database to get the current search path
func (c *Client) getCurrentSearchPath() ([]string, error) {
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

// apply postgres escaping to search path and remove whitespace
func escapeSearchPath(searchPath []string) []string {
	res := make([]string, len(searchPath))
	for idx, path := range searchPath {
		res[idx] = PgEscapeName(strings.TrimSpace(path))
	}
	return res
}
