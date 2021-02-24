package db

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/turbot/steampipe/connection_config"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/utils"
)

// Client wraps over `sql.DB` and gives an interface to the database
type Client struct {
	dbClient       *sql.DB
	schemaMetadata *schema.Metadata
	connectionMap  *connection_config.ConnectionMap
}

var clientSingleton *Client
var cMux sync.Mutex

// Close closes the connection to the database and shuts down the backend
func (c *Client) close() {
	if c.dbClient != nil {
		c.dbClient.Close()
	}

	// set this to nil, so that we recreate stuff the
	// next time GetClient is called
	clientSingleton = nil
}

// GetClient ensures that the database instance is running
// and returns a `Client` to interact with it
func GetClient(autoRefreshConnections bool) (*Client, error) {
	cMux.Lock()
	defer cMux.Unlock()
	if clientSingleton == nil {
		db, err := createSteampipeDbClient()
		if err != nil {
			return nil, err
		}
		clientSingleton = new(Client)
		clientSingleton.dbClient = db
		clientSingleton.loadSchema()

		if autoRefreshConnections {
			RefreshConnections(clientSingleton)
			refreshFunctions(clientSingleton)
		}

		// load the connection state and cache it!
		connectionMap, err := connection_config.GetConnectionState(clientSingleton.schemaMetadata.GetSchemas())
		if err != nil {
			return nil, err
		}
		clientSingleton.connectionMap = &connectionMap
	}

	return clientSingleton, nil
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
func (c *Client) ConnectionMap() *connection_config.ConnectionMap {
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

func (c *Client) setSearchPath() {
	// set the search_path to the available foreign schemas
	// we need to do this here, since postgres resets the search_path on every load.
	schemas := c.schemaMetadata.GetSchemas()

	if len(schemas) > 0 {
		// sort the schema names
		sort.Strings(schemas)
		// set this before the `public` schema gets added
		c.schemaMetadata.SearchPath = schemas
		// add the public schema as the first schema in the search_path. This makes it
		// easier for users to build and work with their own tables, and since it's normally
		// empty, doesn't make using steampipe tables any more difficult.
		schemas = append([]string{"public"}, schemas...)
		// add 'internal' schema as last schema in the search path
		schemas = append(schemas, constants.FunctionSchema)

		schemas = append(schemas, "select")

		// escape the schema names
		escapedSchemas := []string{}

		for _, schema := range schemas {
			escapedSchemas = append(escapedSchemas, PgEscapeName(schema))
		}

		log.Println("[TRACE] setting search path to", schemas)
		query := fmt.Sprintf(
			"alter user %s set search_path to %s;",
			constants.DatabaseUser,
			strings.Join(escapedSchemas, ","),
		)
		c.ExecuteSync(query)
	}
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
			(
				COALESCE(
					-- coalesce to set to '' if NULL
					(
						SELECT 
							pg_catalog.col_description(
								c.oid, cols.ordinal_position :: int
							) 
						FROM 
							pg_catalog.pg_class c 
						WHERE 
							c.relname = cols.table_name
						AND 
							c.relnamespace = (
								SELECT 
									oid 
								FROM 
									pg_catalog.pg_namespace 
								WHERE 
									nspname = cols.table_schema
							)
					),
					''
				)
			) as column_comment, 
			(
				COALESCE(
					-- coalesce to set to '' if NULL
					(
						SELECT 
							pg_catalog.obj_description(c.oid) 
						FROM 
							pg_catalog.pg_class c 
						WHERE 
							c.relname = cols.table_name
						AND 
							c.relnamespace = (
								SELECT 
									oid 
								FROM 
									pg_catalog.pg_namespace 
								WHERE 
									nspname = cols.table_schema
							)
					),
					''
				)
			) as table_comment 
		FROM 
			information_schema.columns cols 
		WHERE 
			cols.table_name in (
				SELECT 
					foreign_table_name 
				FROM 
					information_schema.foreign_tables
			) 
		ORDER BY 
			cols.table_name;
`

	return c.dbClient.Query(query)
}
