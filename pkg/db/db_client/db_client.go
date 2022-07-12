package db_client

import (
	"context"
	"database/sql"
	"log"
	"sort"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants/runtime"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/schema"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"golang.org/x/sync/semaphore"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
)

// DbClient wraps over `sql.DB` and gives an interface to the database
type DbClient struct {
	connectionString          string
	ensureSessionFunc         db_common.EnsureSessionStateCallback
	dbClient                  *sql.DB
	requiredSessionSearchPath []string

	// concurrency management for db session access
	parallelSessionInitLock *semaphore.Weighted

	// a wait group which lets others wait for any running DBSession init to complete
	sessionInitWaitGroup *sync.WaitGroup

	// map of database sessions, keyed to the backend_pid in postgres
	// used to track database sessions that were created
	sessions map[uint32]*db_common.DatabaseSession
	// allows locked access to the 'sessions' map
	sessionsMutex *sync.Mutex

	// list of connection schemas
	foreignSchemaNames []string
	// if a custom search path or a prefix is used, store it here
	customSearchPath []string
	searchPathPrefix []string
	// a cached copy of (viper.GetBool(constants.ArgTiming) && viper.GetString(constants.ArgOutput) == constants.OutputFormatTable)
	// (cached to avoid concurrent access error on viper)
	showTimingFlag bool
	// disable timing - set whilst in process of querying the timing
	disableTiming bool
}

func NewDbClient(ctx context.Context, connectionString string) (*DbClient, error) {
	utils.LogTime("db_client.NewDbClient start")
	defer utils.LogTime("db_client.NewDbClient end")

	db, err := establishConnection(ctx, connectionString)

	if err != nil {
		return nil, err
	}
	client := &DbClient{
		dbClient: db,
		// a waitgroup to keep track of active session initializations
		// so that we don't try to shutdown while an init is underway
		sessionInitWaitGroup: &sync.WaitGroup{},
		// a weighted semaphore to control the maximum number parallel
		// initializations under way
		parallelSessionInitLock: semaphore.NewWeighted(constants.MaxParallelClientInits),
		sessions:                make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex:           &sync.Mutex{},
	}

	// read timing from viper
	client.setShouldShowTiming()

	client.connectionString = connectionString

	// populate foreign schema names - this wil be updated whenever we acquire a session or refresh connections
	if err := client.LoadForeignSchemaNames(ctx); err != nil {
		client.Close(ctx)
		return nil, err
	}

	return client, nil
}

func (c *DbClient) setShouldShowTiming() {
	c.showTimingFlag = viper.GetBool(constants.ArgTiming) && viper.GetString(constants.ArgOutput) == constants.OutputFormatTable
}
func (c *DbClient) shouldShowTiming() bool {
	return c.showTimingFlag && !c.disableTiming
}

func establishConnection(ctx context.Context, connStr string) (*sql.DB, error) {
	utils.LogTime("db_client.establishConnection start")
	defer utils.LogTime("db_client.establishConnection end")

	connConfig, _ := pgx.ParseConfig(connStr)
	connConfig.RuntimeParams = map[string]string{
		// set an app name so that we can track connections from this execution
		"application_name": runtime.PgClientAppName,
	}
	connStr = stdlib.RegisterConnConfig(connConfig)

	db, err := sql.Open("pgx", connStr)
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

	if err := db_common.WaitForConnection(ctx, db); err != nil {
		return nil, err
	}
	return db, nil
}

func (c *DbClient) SetEnsureSessionDataFunc(f db_common.EnsureSessionStateCallback) {
	c.ensureSessionFunc = f
}

// Close implements Client
// closes the connection to the database and shuts down the backend
func (c *DbClient) Close(context.Context) error {
	log.Printf("[TRACE] DbClient.Close %v", c.dbClient)
	if c.dbClient != nil {
		c.sessionInitWaitGroup.Wait()

		// clear the map - so that we can't reuse it
		c.sessions = nil
		return c.dbClient.Close()
	}

	return nil
}

func (c *DbClient) ConnectionMap() *steampipeconfig.ConnectionDataMap {
	return &steampipeconfig.ConnectionDataMap{}
}

// ForeignSchemaNames implements Client
func (c *DbClient) ForeignSchemaNames() []string {
	return c.foreignSchemaNames
}

// LoadForeignSchemaNames implements Client
func (c *DbClient) LoadForeignSchemaNames(ctx context.Context) error {
	res, err := c.dbClient.QueryContext(ctx, "SELECT DISTINCT foreign_table_schema FROM information_schema.foreign_tables")
	if err != nil {
		return err
	}
	// clear foreign schemas
	var foreignSchemaNames []string
	var schema string
	for res.Next() {
		if err := res.Scan(&schema); err != nil {
			return err
		}
		// ignore command schema
		if schema != constants.CommandSchema {
			foreignSchemaNames = append(foreignSchemaNames, schema)
		}
	}
	c.foreignSchemaNames = foreignSchemaNames
	return nil
}

// RefreshSessions terminates the current connections and creates a new one - repopulating session data
func (c *DbClient) RefreshSessions(ctx context.Context) *db_common.AcquireSessionResult {
	utils.LogTime("db_client.RefreshSessions start")
	defer utils.LogTime("db_client.RefreshSessions end")

	if err := c.refreshDbClient(ctx); err != nil {
		return &db_common.AcquireSessionResult{Error: err}
	}
	sessionResult := c.AcquireSession(ctx)
	if sessionResult.Session != nil {
		sessionResult.Session.Close(utils.IsContextCancelled(ctx))
	}
	return sessionResult
}

// refreshDbClient terminates the current connection and opens up a new connection to the service.
func (c *DbClient) refreshDbClient(ctx context.Context) error {
	utils.LogTime("db_client.refreshDbClient start")
	defer utils.LogTime("db_client.refreshDbClient end")

	// wait for any pending inits to finish
	c.sessionInitWaitGroup.Wait()

	// close the connection
	err := c.dbClient.Close()
	if err != nil {
		return err
	}
	db, err := establishConnection(ctx, c.connectionString)
	if err != nil {
		return err
	}
	c.dbClient = db

	return nil
}

// RefreshConnectionAndSearchPaths implements Client
func (c *DbClient) RefreshConnectionAndSearchPaths(ctx context.Context) *steampipeconfig.RefreshConnectionResult {
	// base db client does not refresh connections, it just sets search path
	// (only local db client refreshed connections)
	res := &steampipeconfig.RefreshConnectionResult{}
	if err := c.SetRequiredSessionSearchPath(ctx); err != nil {
		res.Error = err
	}
	return res
}

// GetSchemaFromDB requests for all columns of tables backed by steampipe plugins
// and creates golang struct representations from the result
func (c *DbClient) GetSchemaFromDB(ctx context.Context) (*schema.Metadata, error) {
	utils.LogTime("db_client.GetSchemaFromDB start")
	defer utils.LogTime("db_client.GetSchemaFromDB end")
	connection, err := c.dbClient.Conn(ctx)
	utils.FailOnError(err)

	query := c.buildSchemasQuery()

	tablesResult, err := connection.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	metadata, err := db_common.BuildSchemaMetadata(tablesResult)
	if err != nil {
		return nil, err
	}
	connection.Close()

	searchPath, err := c.GetCurrentSearchPath(ctx)
	if err != nil {
		return nil, err
	}
	metadata.SearchPath = searchPath

	return metadata, nil
}

// GetDefaultSearchPath builds default search path from the connection schemas, book-ended with public and internal
func (c *DbClient) GetDefaultSearchPath(ctx context.Context) []string {
	// get foreign schema names
	searchPath := c.foreignSchemaNames

	sort.Strings(searchPath)
	// add the 'public' schema as the first schema in the search_path. This makes it
	// easier for users to build and work with their own tables, and since it's normally
	// empty, doesn't make using steampipe tables any more difficult.
	searchPath = append([]string{"public"}, searchPath...)
	// add 'internal' schema as last schema in the search path
	searchPath = append(searchPath, constants.FunctionSchema)

	return searchPath
}

func (c *DbClient) buildSchemasQuery() string {
	query := `
WITH distinct_schema AS (
	SELECT DISTINCT(foreign_table_schema) 
	FROM 
		information_schema.foreign_tables 
	WHERE 
		foreign_table_schema <> 'steampipe_command'
)
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
	cols.table_schema in (select * from distinct_schema)
	OR
    LEFT(cols.table_schema,8) = 'pg_temp_'

`
	return query
}
