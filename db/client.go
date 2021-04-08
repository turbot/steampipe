package db

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
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

func (c *Client) setClientSearchPath() error {
	var searchPath []string

	if viper.IsSet("search-path") {
		searchPath = viper.GetStringSlice("search-path")
	} else {
		searchPath = c.schemaMetadata.GetSchemas()
		sort.Strings(searchPath)
	}
	if viper.IsSet("search-path-prefix") {
		prefixedSearchPath := viper.GetStringSlice("search-path-prefix")
		for _, p := range searchPath {
			if !helpers.StringSliceContains(prefixedSearchPath, p) {
				prefixedSearchPath = append(prefixedSearchPath, p)
			}
		}
		searchPath = prefixedSearchPath
	}

	// add the public schema as the first schema in the search_path. This makes it
	// easier for users to build and work with their own tables, and since it's normally
	// empty, doesn't make using steampipe tables any more difficult.
	searchPath = append([]string{"public"}, searchPath...)
	// add 'internal' schema as last schema in the search path
	searchPath = append(searchPath, constants.FunctionSchema)

	// escape the names
	for idx, path := range searchPath {
		searchPath[idx] = PgEscapeName(path)
	}
	q := fmt.Sprintf("set search_path to %s", strings.Join(searchPath, ","))
	_, err := c.ExecuteSync(q)

	if err != nil {
		return err
	}

	c.schemaMetadata.SearchPath = searchPath
	return nil
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
