package db_client

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/steampipeconfig"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/utils"
)

// DbClient wraps over `sql.DB` and gives an interface to the database
type DbClient struct {
	connectionString  string
	ensureSessionFunc db_common.EnsureSessionStateCallback
	dbClient          *sql.DB
	schemaMetadata    *schema.Metadata
}

func NewDbClient(connectionString string) (*DbClient, error) {
	utils.LogTime("db.NewLocalClient start")
	defer utils.LogTime("db.NewLocalClient end")
	db, err := establishConnection(connectionString)
	if err != nil {
		return nil, err
	}
	client := &DbClient{
		dbClient: db,
		// set up a blank struct for the schema metadata
		schemaMetadata: schema.NewMetadata(),
	}
	client.connectionString = connectionString
	client.LoadSchema()

	return client, nil
}

func establishConnection(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	// limit to a single connection as we rely on session scoped data - temp tables and prepared statements
	db.SetMaxOpenConns(1)
	if db_common.WaitForConnection(db) {
		return db, nil
	}
	return nil, fmt.Errorf("could not establish connection")
}

func (c *DbClient) SetEnsureSessionStateFunc(f db_common.EnsureSessionStateCallback) {
	c.ensureSessionFunc = f
}

// Close implements Client
// closes the connection to the database and shuts down the backend
func (c *DbClient) Close() error {
	if c.dbClient != nil {
		return c.dbClient.Close()
	}

	return nil
}

// SchemaMetadata implements Client
// return the latest schema metadata
func (c *DbClient) SchemaMetadata() *schema.Metadata {
	return c.schemaMetadata
}

func (c *DbClient) ConnectionMap() *steampipeconfig.ConnectionDataMap {
	// TODO how to we get connections for remote DB
	return &steampipeconfig.ConnectionDataMap{}
}

// LoadSchema implements Client
// retrieve both the raw query result and a sanitised version in list form
func (c *DbClient) LoadSchema() {
	utils.LogTime("db.LoadSchema start")
	defer utils.LogTime("db.LoadSchema end")

	tablesResult, err := c.getSchemaFromDB()
	utils.FailOnError(err)

	defer tablesResult.Close()

	metadata, err := db_common.BuildSchemaMetadata(tablesResult)
	utils.FailOnError(err)

	c.schemaMetadata.Schemas = metadata.Schemas
	c.schemaMetadata.TemporarySchemaName = metadata.TemporarySchemaName
}

// refreshDbClient terminates the current connection and opens up a new connection to the service.
func (c *DbClient) refreshDbClient(ctx context.Context) error {
	err := c.dbClient.Close()
	if err != nil {
		return err
	}
	db, err := establishConnection(c.connectionString)
	if err != nil {
		return err
	}
	c.dbClient = db

	// setup the session data - prepared statements and introspection tables
	c.ensureServiceState(ctx)

	return nil
}

func (c *DbClient) ensureServiceState(ctx context.Context) error {
	if c.ensureSessionFunc != nil {
		return c.ensureSessionFunc(ctx, c)
	}
	return nil
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

// RefreshConnectionAndSearchPaths implements Client
func (c *DbClient) RefreshConnectionAndSearchPaths() *db_common.RefreshConnectionResult {
	// base db client does not refresh connections, it just sets search path
	// (only local db client refreshed connections)
	res := &db_common.RefreshConnectionResult{}
	if err := c.SetSessionSearchPath(); err != nil {
		res.Error = err
	}
	return res
}
