package db

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/turbot/go-kit/helpers"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

// Client wraps over `sql.DB` and gives an interface to the database
type Client struct {
	dbClient       *sql.DB
	schemaMetadata *schema.Metadata
	connectionMap  *steampipeconfig.ConnectionMap
}

// Close closes the connection to the database and shuts down the backend
func (c *Client) close() {
	if c.dbClient != nil {
		c.dbClient.Close()
	}
}

// GetClient ensures that the database instance is running
// and returns a `Client` to interact with it
func GetClient(autoRefreshConnections bool) (*Client, error) {
	db, err := createSteampipeDbClient()
	if err != nil {
		return nil, err
	}
	client := new(Client)
	client.dbClient = db
	client.loadSchema()

	if autoRefreshConnections {
		client.RefreshConnections()
		refreshFunctions()
	}

	// load the connection state and cache it!
	connectionMap, err := steampipeconfig.GetConnectionState(client.schemaMetadata.GetSchemas())
	if err != nil {
		return nil, err
	}
	client.connectionMap = &connectionMap
	if err := client.setClientSearchPath(); err != nil {
		utils.ShowError(err)
	}

	return client, nil
}

func createSteampipeDbClient() (*sql.DB, error) {
	return createDbClient(constants.DatabaseName, constants.DatabaseUser)
}

func createSteampipeRootDbClient() (*sql.DB, error) {
	return createDbClient(constants.DatabaseName, constants.DatabaseSuperUser)
}

func createPostgresDbClient() (*sql.DB, error) {
	return createDbClient("postgres", constants.DatabaseSuperUser)
}

// set the search path for this client
// if either a search-path or search-path-prexif is set in config, set the seatch path
func (c *Client) setClientSearchPath() error {
	searchPath := viper.GetStringSlice(constants.ArgSearchPath)
	searchPathPrefix := viper.GetStringSlice(constants.ArgSearchPathPrefix)

	// if neither search-path or search-path-prefix are set in config, we have nothing to do
	// - we can just fall back to using th eservice search path
	if len(searchPath) == 0 && len(searchPathPrefix) == 0 {
		return nil
	}

	// if a search path was passed, add 'internal' to the end
	if len(searchPath) > 0 {
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.FunctionSchema)
	} else {
		// so a prefix was set, but no search path
		// in this case we need to load the existing service search path
		// (NOTE: we cannot just build a default search path from schemas,
		// as an argument may have been passed to service start to set the service search path)
		searchPath, _ = c.getCurrentSearchPath()
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

// set the search path for the db service (by setting it on the steampipe user)
func (c *Client) setServiceSearchPath() error {
	// set the search_path to the available foreign schemas
	// or the one set by the user in config
	var searchPath []string

	// since this is the service starting up, use the ConfigKeyDatabaseSearchPath config
	// (this is the value specified in the database config)
	if viper.IsSet(constants.ConfigKeyDatabaseSearchPath) {
		searchPath = viper.GetStringSlice(constants.ConfigKeyDatabaseSearchPath)
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.FunctionSchema)
	} else {
		// so no search path was set in config - build a search path from the connection schemas,
		// and add in public and internal
		searchPath = c.getDefaultSearchPath(searchPath)
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

// build default search path from the connection schemas, bookended with public and internal
func (c *Client) getDefaultSearchPath(searchPath []string) []string {
	searchPath = c.schemaMetadata.GetSchemas()
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

// escape search path and remove whitespace
// NOTE: this mutates search path
func escapeSearchPath(searchPath []string) []string {
	res := make([]string, len(searchPath))
	for idx, path := range searchPath {
		res[idx] = PgEscapeName(strings.TrimSpace(path))
	}
	return res
}

func createDbClient(dbname string, username string) (*sql.DB, error) {

	log.Println("[TRACE] createDbClient")
	info, err := GetStatus()

	if err != nil {
		return nil, err
	}

	if info == nil {
		return nil, fmt.Errorf("steampipe service is not running")
	}

	// Connect to the database using the first listen address, which is usually localhost
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", info.Listen[0], info.Port, username, dbname)

	log.Println("[TRACE] status: ", info)
	log.Println("[TRACE] Connection string: ", psqlInfo)

	// connect to the database using the postgres driver
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if waitForConnection(db) {
		return db, nil
	}

	return nil, fmt.Errorf("could not establish connection with database")
}

// waits for the db to start accepting connections and returns true
// returns false if the dbClient does not start within a stipulated time,
func waitForConnection(conn *sql.DB) bool {
	timeoutAt := time.Now().Add(5 * time.Second)
	intervalMs := 10

	for {
		pingErr := conn.Ping()
		if pingErr == nil {
			return true
		}
		if timeoutAt.Before(time.Now()) {
			break
		}
		time.Sleep(time.Duration(intervalMs) * time.Millisecond)
	}
	return false
}

// SchemaMetadata :: returns the latest schema metadata
func (c *Client) SchemaMetadata() *schema.Metadata {
	return c.schemaMetadata
}

// ConnectionMap :: returns the latest connection map
func (c *Client) ConnectionMap() *steampipeconfig.ConnectionMap {
	return c.connectionMap
}

// return both the raw query result and a sanitised version in list form
func (c *Client) loadSchema() {
	tablesResult, err := c.getSchemaFromDB()
	utils.FailOnError(err)

	defer tablesResult.Close()

	c.schemaMetadata, err = buildSchemaMetadata(tablesResult)
	utils.FailOnError(err)
}

func (c *Client) getSchemaFromDB() (*sql.Rows, error) {
	query := `
		SELECT 
			table_name, 
			column_name,
			column_default,
			is_nullable,
			data_type,
			table_schema,
			(COALESCE(pg_catalog.col_description(c.oid, cols.ordinal_position :: int),'')) as column_comment,
			(COALESCE(pg_catalog.obj_description(c.oid),'')) as table_comment
		FROM
			information_schema.columns cols
		LEFT JOIN
			pg_catalog.pg_namespace nsp ON nsp.nspname = cols.table_schema
		LEFT JOIN
			pg_catalog.pg_class c ON c.relname = cols.table_name AND c.relnamespace = nsp.oid
		WHERE
			cols.table_name in (
				SELECT 
					foreign_table_name 
				FROM 
					information_schema.foreign_tables
			) 
			OR
			cols.table_schema = 'public'
		ORDER BY 
			cols.table_schema, cols.table_name, cols.column_name;
`

	return c.dbClient.Query(query)
}
