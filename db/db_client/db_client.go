package db_client

import (
	"database/sql"
	"fmt"

	"github.com/turbot/steampipe/steampipeconfig"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/utils"
)

// DbClient wraps over `sql.DB` and gives an interface to the database
type DbClient struct {
	dbClient       *sql.DB
	schemaMetadata *schema.Metadata
}

func NewDbClient(connectionString string) (*DbClient, error) {
	utils.LogTime("db.NewLocalClient start")
	defer utils.LogTime("db.NewLocalClient end")

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	// limit to a single connection as we rely on session scoped data - temp tables and prepared statements
	db.SetMaxOpenConns(1)
	return NewDbClientFromSqlClient(db)
}

// NewDbClientFromSqlClient creates a DbClient from an existing sql.DB client
// used by LocalDbClient
func NewDbClientFromSqlClient(db *sql.DB) (*DbClient, error) {
	client := &DbClient{
		dbClient: db,
		// set up a blank struct for the schema metadata
		schemaMetadata: schema.NewMetadata(),
	}
	client.LoadSchema()

	return client, nil
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
// base db client does not refresh connections, it just sets searhc path
// (only local db client refreshed connections)
func (c *DbClient) RefreshConnectionAndSearchPaths() *db_common.RefreshConnectionResult {
	res := &db_common.RefreshConnectionResult{}
	if err := c.SetSessionSearchPath(); err != nil {
		res.Error = err
	}
	return res
}
