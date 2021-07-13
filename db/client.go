package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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
func (c *Client) Close() error {
	if c.dbClient != nil {
		return c.dbClient.Close()
	}
	return nil
}

// NewClient ensures that the database instance is running
// and returns a `Client` to interact with it
func NewClient(autoRefreshConnections bool) (*Client, *RefreshConnectionResult) {
	utils.LogTime("db.NewClient start")
	defer utils.LogTime("db.NewClient end")
	refreshResult := &RefreshConnectionResult{}
	db, err := createSteampipeDbClient()
	if err != nil {
		refreshResult.Error = err
		return nil, refreshResult
	}
	client := new(Client)
	client.dbClient = db

	// setup a blank struct for the schema metadata
	client.schemaMetadata = schema.NewMetadata()

	client.loadSchema()

	if autoRefreshConnections {
		refreshResult = client.RefreshConnections()
		if refreshResult.Error != nil {
			client.Close()
			return nil, refreshResult
		}

		if err := refreshFunctions(); err != nil {
			client.Close()
			refreshResult.Error = err
			return nil, refreshResult
		}
	}

	// update the service and client search paths
	if err := client.SetClientSearchPath(); err != nil {
		refreshResult.Error = err
		return nil, refreshResult
	}
	// TODO 0712 also set service search path?
	if err := client.SetServiceSearchPath(); err != nil {
		refreshResult.Error = err
		return nil, refreshResult
	}

	// finally update the connection map
	if err = client.updateConnectionMap(); err != nil {
		refreshResult.Error = err
		return nil, refreshResult
	}

	return client, refreshResult
}

func createSteampipeDbClient() (*sql.DB, error) {
	utils.LogTime("db.createSteampipeDbClient start")
	defer utils.LogTime("db.createSteampipeDbClient end")

	return createDbClient(constants.DatabaseName, constants.DatabaseUser)
}

// close and reopen db client
func (c *Client) refreshDbClient() error {
	c.dbClient.Close()
	db, err := createSteampipeDbClient()
	if err != nil {
		return err
	}
	c.dbClient = db
	return nil
}

func createSteampipeRootDbClient() (*sql.DB, error) {
	return createDbClient(constants.DatabaseName, constants.DatabaseSuperUser)
}

func createPostgresDbClient() (*sql.DB, error) {
	return createDbClient("postgres", constants.DatabaseSuperUser)
}

func createDbClient(dbname string, username string) (*sql.DB, error) {
	utils.LogTime("db.createDbClient start")
	defer utils.LogTime("db.createDbClient end")

	log.Println("[TRACE] createDbClient")
	info, err := GetStatus()

	if err != nil {
		return nil, err
	}

	if info == nil {
		return nil, fmt.Errorf("steampipe service is not running")
	}

	// Connect to the database using the first listen address, which is usually localhost
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", info.Listen[0], info.Port, username, dbname, SslMode())

	log.Println("[TRACE] status: ", info)
	log.Println("[TRACE] Connection string: ", psqlInfo)

	// connect to the database using the postgres driver
	utils.LogTime("db.createDbClient connection open start")
	db, err := sql.Open("postgres", psqlInfo)
	db.SetMaxOpenConns(1)
	utils.LogTime("db.createDbClient connection open end")

	if err != nil {
		return nil, err
	}

	if waitForConnection(db) {
		return db, nil
	}

	return nil, fmt.Errorf("could not establish connection with database")
}

func executeSqlAsRoot(statements []string) ([]sql.Result, error) {
	var results []sql.Result
	rootClient, err := createSteampipeRootDbClient()
	if err != nil {
		return nil, err
	}
	defer func() {
		rootClient.Close()
	}()

	for _, statement := range statements {
		result, err := rootClient.Exec(statement)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// waits for the db to start accepting connections and returns true
// returns false if the dbClient does not start within a stipulated time,
func waitForConnection(conn *sql.DB) bool {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	pingTimer := time.NewTicker(10 * time.Millisecond)
	timeoutAt := time.After(5 * time.Second)
	defer pingTimer.Stop()
	for {
		select {
		case <-pingTimer.C:
			pingErr := conn.Ping()
			if pingErr == nil {
				return true
			}
		case <-timeoutAt:
			return false
		}
	}
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
	utils.LogTime("db.loadSchema start")
	defer utils.LogTime("db.loadSchema end")

	tablesResult, err := c.getSchemaFromDB()
	utils.FailOnError(err)

	defer tablesResult.Close()

	metadata, err := buildSchemaMetadata(tablesResult)
	utils.FailOnError(err)

	c.schemaMetadata.Schemas = metadata.Schemas
	c.schemaMetadata.TemporarySchemaName = metadata.TemporarySchemaName
}

func (c *Client) getSchemaFromDB() (*sql.Rows, error) {
	utils.LogTime("db.getSchemaFromDB start")
	defer utils.LogTime("db.getSchemaFromDB end")

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
			OR
			LEFT(cols.table_schema,8) = 'pg_temp_'
		ORDER BY 
			cols.table_schema, cols.table_name, cols.column_name;
`

	return c.dbClient.Query(query)
}
