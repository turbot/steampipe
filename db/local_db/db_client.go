package local_db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

// DbClient wraps over `sql.DB` and gives an interface to the database
type DbClient struct {
	dbClient         *sql.DB
	schemaMetadata   *schema.Metadata
	connectionMap    *steampipeconfig.ConnectionMap
	invoker          constants.Invoker
	connectionString string
}

// Close closes the connection to the database and shuts down the backend
func (c *DbClient) Close() error {
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
func NewLocalClient(invoker constants.Invoker) (*DbClient, error) {
	utils.LogTime("db.NewLocalClient start")
	defer utils.LogTime("db.NewLocalClient end")

	db, err := createSteampipeDbClient()
	if err != nil {
		return nil, err
	}
	client := &DbClient{
		invoker:  invoker,
		dbClient: db,
		// setup a blank struct for the schema metadata
		schemaMetadata: schema.NewMetadata(),
	}

	client.LoadSchema()

	return client, nil
}

func NewDbClient(invoker constants.Invoker, connectionString string) (*DbClient, error) {
	utils.LogTime("db.NewLocalClient start")
	defer utils.LogTime("db.NewLocalClient end")

	db, err := createDbClientWithConnectionString(connectionString)
	if err != nil {
		return nil, err
	}
	client := &DbClient{
		invoker:  invoker,
		dbClient: db,
		// setup a blank struct for the schema metadata
		schemaMetadata: schema.NewMetadata(),
	}
	client.LoadSchema()

	return client, nil
}

func (c *DbClient) IsLocal() bool {
	return c.connectionString == ""
}

// SchemaMetadata returns the latest schema metadata
func (c *DbClient) SchemaMetadata() *schema.Metadata {
	return c.schemaMetadata
}

// ConnectionMap returns the latest connection map
func (c *DbClient) ConnectionMap() *steampipeconfig.ConnectionMap {
	return c.connectionMap
}

// LoadSchema retrieves both the raw query result and a sanitised version in list form
func (c *DbClient) LoadSchema() {
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

func executeSqlAsRoot(statements ...string) ([]sql.Result, error) {
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

func (c *DbClient) getSchemaFromDB() (*sql.Rows, error) {
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
		    (
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
			)
			AND
			cols.table_schema <> '%s'
		ORDER BY 
			cols.table_schema, cols.table_name, cols.column_name;
`
	// we do NOT want to fetch the command schema
	return c.dbClient.Query(fmt.Sprintf(query, constants.CommandSchema))
}
