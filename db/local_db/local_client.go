package local_db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/turbot/steampipe/db/db_common"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

// LocalClient wraps over `sql.DB` and gives an interface to the database
type LocalClient struct {
	dbClient       *sql.DB
	schemaMetadata *schema.Metadata
	connectionMap  *steampipeconfig.ConnectionMap
	invoker        constants.Invoker
}

// Close closes the connection to the database and shuts down the backend
func (c *LocalClient) Close() error {
	if c.dbClient != nil {
		if err := c.dbClient.Close(); err != nil {
			return err
		}
	}
	ShutdownService(c.invoker)
	return nil
}

// NewLocalClient ensures that the database instance is running
// and returns a `Client` to interact with it
func NewLocalClient(invoker constants.Invoker) (*LocalClient, error) {
	utils.LogTime("db.NewLocalClient start")
	defer utils.LogTime("db.NewLocalClient end")

	db, err := createSteampipeDbClient()
	if err != nil {
		return nil, err
	}
	client := &LocalClient{
		invoker: invoker,
	}
	client.dbClient = db

	// setup a blank struct for the schema metadata
	client.schemaMetadata = schema.NewMetadata()

	client.LoadSchema()

	return client, nil
}

func (c *LocalClient) RefreshConnectionAndSearchPaths() *db_common.RefreshConnectionResult {
	res := c.RefreshConnections()
	if res.Error != nil {
		return res
	}
	if err := refreshFunctions(); err != nil {
		res.Error = err
		return res
	}

	// load the connection state and cache it!
	connectionMap, err := steampipeconfig.GetConnectionState(c.schemaMetadata.GetSchemas())
	if err != nil {
		res.Error = err
		return res
	}
	c.connectionMap = &connectionMap
	// set service search path first - client may fall back to using it
	if err := c.SetServiceSearchPath(); err != nil {
		res.Error = err
		return res
	}
	if err := c.SetClientSearchPath(); err != nil {
		res.Error = err
		return res
	}

	return res
}

// SchemaMetadata returns the latest schema metadata
func (c *LocalClient) SchemaMetadata() *schema.Metadata {
	return c.schemaMetadata
}

// ConnectionMap returns the latest connection map
func (c *LocalClient) ConnectionMap() *steampipeconfig.ConnectionMap {
	return c.connectionMap
}

// LoadSchema retrieves both the raw query result and a sanitised version in list form
// todo share this between local and remote client - make a function not a method??
func (c *LocalClient) LoadSchema() {
	utils.LogTime("db.LoadSchema start")
	defer utils.LogTime("db.LoadSchema end")

	tablesResult, err := c.getSchemaFromDB()
	utils.FailOnError(err)

	defer tablesResult.Close()

	metadata, err := buildSchemaMetadata(tablesResult)
	utils.FailOnError(err)

	c.schemaMetadata.Schemas = metadata.Schemas
	c.schemaMetadata.TemporarySchemaName = metadata.TemporarySchemaName
}

// close and reopen db client
func (c *LocalClient) refreshDbClient() error {
	c.dbClient.Close()
	db, err := createSteampipeDbClient()
	if err != nil {
		return err
	}
	c.dbClient = db
	return nil
}

func createSteampipeDbClient() (*sql.DB, error) {
	utils.LogTime("db.createSteampipeDbClient start")
	defer utils.LogTime("db.createSteampipeDbClient end")

	return createDbClient(constants.DatabaseName, constants.DatabaseUser)
}

func createRootDbClient() (*sql.DB, error) {
	utils.LogTime("db.createSteampipeRootDbClient start")
	defer utils.LogTime("db.createSteampipeRootDbClient end")

	return createDbClient(constants.DatabaseName, constants.DatabaseSuperUser)
}

func createDbClient(dbname string, username string) (*sql.DB, error) {
	utils.LogTime("db.createDbClient start")
	utils.LogTime(fmt.Sprintf("to %s with %s", dbname, username))
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
	rootClient, err := createRootDbClient()
	if err != nil {
		return nil, err
	}
	defer rootClient.Close()

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

func (c *LocalClient) getSchemaFromDB() (*sql.Rows, error) {
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
