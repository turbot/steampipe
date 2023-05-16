package db_client

import (
	"context"
	"fmt"
	"golang.org/x/exp/maps"
	"log"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"golang.org/x/sync/semaphore"
)

// DbClient wraps over `sql.DB` and gives an interface to the database
type DbClient struct {
	connectionString string
	pool             *pgxpool.Pool

	// concurrency management for db session access
	parallelSessionInitLock *semaphore.Weighted

	// map of database sessions, keyed to the backend_pid in postgres
	// used to update session search path where necessary
	sessions map[uint32]*db_common.DatabaseSession
	// allows locked access to the 'sessions' map
	sessionsMutex *sync.Mutex

	// if a custom search path or a prefix is used, store it here
	customSearchPath []string
	searchPathPrefix []string
	// the default user search path
	userSearchPath []string
	// a cached copy of (viper.GetBool(constants.ArgTiming) && viper.GetString(constants.ArgOutput) == constants.OutputFormatTable)
	// (cached to avoid concurrent access error on viper)
	showTimingFlag bool
	// disable timing - set whilst in process of querying the timing
	disableTiming        bool
	onConnectionCallback DbConnectionCallback
}

func NewDbClient(ctx context.Context, connectionString string, onConnectionCallback DbConnectionCallback) (*DbClient, error) {
	utils.LogTime("db_client.NewDbClient start")
	defer utils.LogTime("db_client.NewDbClient end")

	wg := &sync.WaitGroup{}
	// wrap onConnectionCallback to use wait group
	var wrappedOnConnectionCallback DbConnectionCallback
	if onConnectionCallback != nil {
		wrappedOnConnectionCallback = func(ctx context.Context, conn *pgx.Conn) error {
			wg.Add(1)
			defer wg.Done()
			return onConnectionCallback(ctx, conn)
		}
	}

	client := &DbClient{
		// a weighted semaphore to control the maximum number parallel
		// initializations under way
		parallelSessionInitLock: semaphore.NewWeighted(constants.MaxParallelClientInits),
		sessions:                make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex:           &sync.Mutex{},
		// store the callback
		onConnectionCallback: wrappedOnConnectionCallback,
		connectionString:     connectionString,
	}

	if err := client.establishConnectionPool(ctx); err != nil {
		return nil, err
	}

	// set user search path
	err := client.LoadUserSearchPath(ctx)
	if err != nil {
		return nil, err
	}

	// populate customSearchPath
	if err := client.SetRequiredSessionSearchPath(ctx); err != nil {
		client.Close(ctx)
		return nil, err
	}

	return client, nil
}

func (c *DbClient) setShouldShowTiming(ctx context.Context, session *db_common.DatabaseSession) {
	currentShowTimingFlag := viper.GetBool(constants.ArgTiming)

	// if we are turning timing ON, fetch the ScanMetadataMaxId
	// to ensure we only select the relevant scan metadata table entries
	if currentShowTimingFlag && !c.showTimingFlag {
		c.updateScanMetadataMaxId(ctx, session)
	}

	c.showTimingFlag = currentShowTimingFlag
}

func (c *DbClient) shouldShowTiming() bool {
	return c.showTimingFlag && !c.disableTiming
}

// Close implements Client
// closes the connection to the database and shuts down the backend
func (c *DbClient) Close(context.Context) error {
	log.Printf("[TRACE] DbClient.Close %v", c.pool)
	if c.pool != nil {
		// clear the sessions map - so that we can't reuse it
		c.sessions = nil
		c.pool.Close()
	}

	return nil
}

// RefreshSessions terminates the current connections and creates a new one - repopulating session data
func (c *DbClient) RefreshSessions(ctx context.Context) (res *db_common.AcquireSessionResult) {
	utils.LogTime("db_client.RefreshSessions start")
	defer utils.LogTime("db_client.RefreshSessions end")

	if err := c.refreshDbClient(ctx); err != nil {
		res.Error = err
		return res
	}
	res = c.AcquireSession(ctx)
	if res.Session != nil {
		res.Session.Close(error_helpers.IsContextCanceled(ctx))
	}
	return res
}

// GetSchemaFromDB  retrieves schemas for all steampipe connections (EXCEPT DISABLED CONNECTIONS)
// NOTE: it optimises the schema extraction by extracting schema information for
// connections backed by distinct plugins and then fanning back out.
func (c *DbClient) GetSchemaFromDB(ctx context.Context) (*db_common.SchemaMetadata, error) {
	conn, _, err := c.GetDatabaseConnectionWithRetries(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// for optimisation purposes, try to loade connection state and build a map of schemas to load
	// (if we are connected to a remote server running an older CLI,
	// this load may fail, in which case bypass the optimisation)
	var schemas []string
	var connectionSchemaMap steampipeconfig.ConnectionSchemaMap
	connectionStateMap, err := steampipeconfig.LoadConnectionState(ctx, conn.Conn(), steampipeconfig.WithWaitUntilLoading())
	// NOTE: if we failed to load conenction state, this may be because we are connected to an older version of the CLI
	// use legacy (v0.19.x) schema loading code
	if err != nil {
		return c.GetSchemaFromDBLegacy(ctx, conn)
	}

	// build a ConnectionSchemaMap object to identify the schemas to load
	connectionSchemaMap = steampipeconfig.NewConnectionSchemaMap(ctx, connectionStateMap, c.GetRequiredSessionSearchPath())
	if err != nil {
		return nil, err
	}

	// get the unique schema - we use this to limit the schemas we load from the database
	schemas = maps.Keys(connectionSchemaMap)

	// build a query to retrieve these schemas
	query := c.buildSchemasQuery(schemas...)

	//execute
	tablesResult, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// build schema metadata from query result
	metadata, err := db_common.BuildSchemaMetadata(tablesResult)
	if err != nil {
		return nil, err
	}

	// we now need to add in all other schemas which have the same schemas as those we have loaded
	for loadedSchema, otherSchemas := range connectionSchemaMap {
		// all 'otherSchema's have the same schema as loadedSchema
		exemplarSchema, ok := metadata.Schemas[loadedSchema]
		if !ok {
			// should can happen in the case of a dynamic plugin with no tables - use empty schema
			exemplarSchema = make(map[string]db_common.TableSchema)
		}

		for _, s := range otherSchemas {
			metadata.Schemas[s] = exemplarSchema
		}
	}

	return metadata, nil
}

func (c *DbClient) GetSchemaFromDBLegacy(ctx context.Context, conn *pgxpool.Conn) (*db_common.SchemaMetadata, error) {
	// build a query to retrieve these schemas
	query := c.buildSchemasQueryLegacy()

	//execute
	tablesResult, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// build schema metadata from query result
	return db_common.BuildSchemaMetadata(tablesResult)
}

// refreshDbClient terminates the current connection and opens up a new connection to the service.
func (c *DbClient) refreshDbClient(ctx context.Context) error {
	utils.LogTime("db_client.refreshDbClient start")
	defer utils.LogTime("db_client.refreshDbClient end")

	// close the connection pool and recreate
	c.pool.Close()
	if err := c.establishConnectionPool(ctx); err != nil {
		return err
	}

	return nil
}

func (c *DbClient) buildSchemasQuery(schemas ...string) string {
	for idx, s := range schemas {
		schemas[idx] = fmt.Sprintf("'%s'", s)
	}

	// build the schemas filter clause
	schemaClause := ""
	if len(schemas) > 0 {
		schemaClause = fmt.Sprintf(`
    cols.table_schema in (%s)
	OR`, strings.Join(schemas, ","))
	}

	query := fmt.Sprintf(`
SELECT
    table_name,
    column_name,
    column_default,
    is_nullable,
    data_type,
	udt_name,
    table_schema,
    (COALESCE(pg_catalog.col_description(c.oid, cols.ordinal_position :: int),'')) as column_comment,
    (COALESCE(pg_catalog.obj_description(c.oid),'')) as table_comment
FROM
    information_schema.columns cols
LEFT JOIN
    pg_catalog.pg_namespace nsp ON nsp.nspname = cols.table_schema
LEFT JOIN
    pg_catalog.pg_class c ON c.relname = cols.table_name AND c.relnamespace = nsp.oid
WHERE %s
	LEFT(cols.table_schema,8) = 'pg_temp_'
`, schemaClause)
	return query
}
func (c *DbClient) buildSchemasQueryLegacy() string {

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
    udt_name,
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
