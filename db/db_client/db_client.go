package db_client

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/steampipeconfig"
	"golang.org/x/sync/semaphore"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/utils"
)

// DbClient wraps over `sql.DB` and gives an interface to the database
type DbClient struct {
	connectionString  string
	ensureSessionFunc db_common.EnsureSessionStateCallback
	dbClient          *sql.DB
	// map of database sessions, keyed to the backend_pid in postgres
	// used to track whether a given session has been initialised
	initializedSessions       SessionStatMap
	schemaMetadata            *schema.Metadata
	requiredSessionSearchPath []string

	// concurrency management for db session access
	sessionMapMutex         *sync.Mutex
	sessionAcquireMutex     *sync.Mutex
	parallelSessionInitLock *semaphore.Weighted
	sessionInitWaitGroup    *sync.WaitGroup
}

func NewDbClient(connectionString string) (*DbClient, error) {
	utils.LogTime("db_client.NewDbClient start")
	defer utils.LogTime("db_client.NewDbClient end")
	db, err := establishConnection(connectionString)
	if err != nil {
		return nil, err
	}
	client := &DbClient{
		dbClient:            db,
		initializedSessions: make(SessionStatMap),
		// set up a blank struct for the schema metadata
		schemaMetadata: schema.NewMetadata(),
		// a waitgroup to keep track of active session initializations
		// so that we don't try to shutdown while an init is underway
		sessionInitWaitGroup:    &sync.WaitGroup{},
		sessionMapMutex:         new(sync.Mutex),
		sessionAcquireMutex:     new(sync.Mutex),
		parallelSessionInitLock: semaphore.NewWeighted(5),
	}
	client.connectionString = connectionString
	client.LoadSchema()

	return client, nil
}

func establishConnection(connStr string) (*sql.DB, error) {
	utils.LogTime("db_client.establishConnection start")
	defer utils.LogTime("db_client.establishConnection end")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	maxParallel := constants.DefaultMaxConnections
	if viper.IsSet(constants.ArgMaxParallel) {
		maxParallel = viper.GetInt(constants.ArgMaxParallel)
	}

	db.SetMaxOpenConns(maxParallel)
	db.SetMaxIdleConns(maxParallel)
	// never close connection even if idle
	db.SetConnMaxIdleTime(0)
	// never close connection because of age
	db.SetConnMaxLifetime(0)

	if err := db_common.WaitForConnection(db); err != nil {
		return nil, err
	}
	return db, nil
}

func (c *DbClient) SetEnsureSessionDataFunc(f db_common.EnsureSessionStateCallback) {
	c.ensureSessionFunc = f
}

// Close implements Client
// closes the connection to the database and shuts down the backend
func (c *DbClient) Close() error {
	if c.dbClient != nil {
		// clear the map - so that we can't reuse it
		log.Printf("[TRACE] Number of unique database sessions: %d\n", len(c.initializedSessions))
		c.initializedSessions = nil
		c.sessionInitWaitGroup.Wait()
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
	utils.LogTime("db_client.LoadSchema start")
	defer utils.LogTime("db_client.LoadSchema end")

	connection, err := c.dbClient.Conn(context.Background())
	utils.FailOnError(err)
	defer connection.Close()

	tablesResult, err := c.getSchemaFromDB(connection)
	utils.FailOnError(err)
	defer tablesResult.Close()

	metadata, err := db_common.BuildSchemaMetadata(tablesResult)
	utils.FailOnError(err)

	c.schemaMetadata.Schemas = metadata.Schemas
	c.schemaMetadata.TemporarySchemaName = metadata.TemporarySchemaName
}

// RefreshSessions terminates the current connections.
func (c *DbClient) RefreshSessions(ctx context.Context) error {
	utils.LogTime("db_client.RefreshSessions start")
	defer utils.LogTime("db_client.RefreshSessions end")

	return c.refreshDbClient(ctx)
}

// refreshDbClient terminates the current connection and opens up a new connection to the service.
func (c *DbClient) refreshDbClient(ctx context.Context) error {
	utils.LogTime("db_client.refreshDbClient start")
	defer utils.LogTime("db_client.refreshDbClient end")

	// clear the initializedSessions map
	c.initializedSessions = make(SessionStatMap)
	err := c.dbClient.Close()
	if err != nil {
		return err
	}
	db, err := establishConnection(c.connectionString)
	if err != nil {
		return err
	}
	c.dbClient = db

	return nil
}

func (c *DbClient) getSchemaFromDB(conn *sql.Conn) (*sql.Rows, error) {
	utils.LogTime("db_client.getSchemaFromDB start")
	defer utils.LogTime("db_client.getSchemaFromDB end")

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
	return conn.QueryContext(context.Background(), fmt.Sprintf(query, constants.CommandSchema))
}

// RefreshConnectionAndSearchPaths implements Client
func (c *DbClient) RefreshConnectionAndSearchPaths() *steampipeconfig.RefreshConnectionResult {
	// base db client does not refresh connections, it just sets search path
	// (only local db client refreshed connections)
	res := &steampipeconfig.RefreshConnectionResult{}
	if err := c.SetSessionSearchPath(); err != nil {
		res.Error = err
	}
	return res
}
