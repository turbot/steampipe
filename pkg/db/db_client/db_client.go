package db_client

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/serversettings"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/semaphore"
)

// DbClient wraps over `sql.DB` and gives an interface to the database
type DbClient struct {
	connectionString string

	// connection userPool for user initiated queries
	userPool *pgxpool.Pool

	// connection used to run system/plumbing queries (connection state, server settings)
	managementPool *pgxpool.Pool

	// the settings of the server that this client is connected to
	serverSettings *db_common.ServerSettings

	// this flag is set if the service that this client
	// is connected to is running in the same physical system
	isLocalService bool

	// concurrency management for db session access
	parallelSessionInitLock *semaphore.Weighted

	// map of database sessions, keyed to the backend_pid in postgres
	// used to update session search path where necessary
	// Session lifecycle: entries are added when connections are established and automatically
	// removed via a pgxpool BeforeClose callback when connections are closed by the pool.
	// This prevents memory accumulation from stale connection entries (see issue #3737)
	sessions map[uint32]*db_common.DatabaseSession

	// allows locked access to the 'sessions' map
	sessionsMutex    *sync.Mutex
	sessionsLockFlag atomic.Bool

	// if a custom search path or a prefix is used, store it here
	customSearchPath []string
	searchPathPrefix []string
	// allows locked access to customSearchPath and searchPathPrefix
	searchPathMutex *sync.RWMutex
	// the default user search path
	userSearchPath []string
	// disable timing - set whilst in process of querying the timing
	disableTiming        atomic.Bool
	onConnectionCallback DbConnectionCallback
}

func NewDbClient(ctx context.Context, connectionString string, opts ...ClientOption) (_ *DbClient, err error) {
	utils.LogTime("db_client.NewDbClient start")
	defer utils.LogTime("db_client.NewDbClient end")

	client := &DbClient{
		// a weighted semaphore to control the maximum number parallel
		// initializations under way
		parallelSessionInitLock: semaphore.NewWeighted(constants.MaxParallelClientInits),
		sessions:                make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex:           &sync.Mutex{},
		searchPathMutex:         &sync.RWMutex{},
		connectionString:        connectionString,
	}

	defer func() {
		if err != nil {
			// try closing the client
			client.Close(ctx)
		}
	}()

	config := clientConfig{}
	for _, o := range opts {
		o(&config)
	}

	if err := client.establishConnectionPool(ctx, config); err != nil {
		return nil, err
	}

	// load up the server settings
	if err := client.loadServerSettings(ctx); err != nil {
		return nil, err
	}

	// set user search path
	if err := client.LoadUserSearchPath(ctx); err != nil {
		return nil, err
	}

	// populate customSearchPath
	if err := client.SetRequiredSessionSearchPath(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *DbClient) closePools() {
	if c.userPool != nil {
		c.userPool.Close()
	}
	if c.managementPool != nil {
		c.managementPool.Close()
	}
}

func (c *DbClient) loadServerSettings(ctx context.Context) error {
	serverSettings, err := serversettings.Load(ctx, c.managementPool)
	if err != nil {
		if notFound := db_common.IsRelationNotFoundError(err); notFound {
			// when connecting to pre-0.21.0 services, the steampipe_server_settings table will not be available.
			// this is expected and not an error
			// code which uses steampipe_server_settings should handle this
			log.Printf("[TRACE] could not find %s.%s table. skipping\n", constants.InternalSchema, constants.ServerSettingsTable)
			return nil
		}
		return err
	}
	c.serverSettings = serverSettings
	log.Println("[TRACE] loaded server settings:", serverSettings)
	return nil
}

func (c *DbClient) shouldFetchTiming() bool {
	// check for override flag (this is to prevent timing being fetched when we read the timing metadata table)
	if c.disableTiming.Load() {
		return false
	}
	// only fetch timing if timing flag is set, or output is JSON
	return (viper.GetString(pconstants.ArgTiming) != pconstants.ArgOff) ||
		(viper.GetString(pconstants.ArgOutput) == constants.OutputFormatJSON)

}
func (c *DbClient) shouldFetchVerboseTiming() bool {
	return (viper.GetString(pconstants.ArgTiming) == pconstants.ArgVerbose) ||
		(viper.GetString(pconstants.ArgOutput) == constants.OutputFormatJSON)
}

// lockSessions acquires the sessionsMutex and tracks ownership for tryLock compatibility.
func (c *DbClient) lockSessions() {
	if c.sessionsMutex == nil {
		return
	}
	c.sessionsLockFlag.Store(true)
	c.sessionsMutex.Lock()
}

// sessionsTryLock attempts to acquire the sessionsMutex without blocking.
// Returns false if the lock is already held.
func (c *DbClient) sessionsTryLock() bool {
	if c.sessionsMutex == nil {
		return false
	}
	// best-effort: only one contender sets the flag and proceeds to lock
	if !c.sessionsLockFlag.CompareAndSwap(false, true) {
		return false
	}
	c.sessionsMutex.Lock()
	return true
}

func (c *DbClient) sessionsUnlock() {
	if c.sessionsMutex == nil {
		return
	}
	c.sessionsMutex.Unlock()
	c.sessionsLockFlag.Store(false)
}

// ServerSettings returns the settings of the steampipe service that this DbClient is connected to
//
// Keep in mind that when connecting to pre-0.21.x servers, the server_settings data is not available. This is expected.
// Code which read server_settings should take this into account.
func (c *DbClient) ServerSettings() *db_common.ServerSettings {
	return c.serverSettings
}

// RegisterNotificationListener has an empty implementation
// NOTE: we do not (currently) support notifications from remote connections
func (c *DbClient) RegisterNotificationListener(func(notification *pgconn.Notification)) {}

// Close implements Client

// closes the connection to the database and shuts down the backend
func (c *DbClient) Close(context.Context) error {
	log.Printf("[TRACE] DbClient.Close %v", c.userPool)
	c.closePools()
	// nullify active sessions, since with the closing of the pools
	// none of the sessions will be valid anymore
	// Acquire mutex to prevent concurrent access to sessions map
	c.lockSessions()
	c.sessions = nil
	c.sessionsUnlock()

	return nil
}

// GetSchemaFromDB  retrieves schemas for all steampipe connections (EXCEPT DISABLED CONNECTIONS)
// NOTE: it optimises the schema extraction by extracting schema information for
// connections backed by distinct plugins and then fanning back out.
func (c *DbClient) GetSchemaFromDB(ctx context.Context) (*db_common.SchemaMetadata, error) {
	log.Printf("[INFO] DbClient GetSchemaFromDB")
	mgmtConn, err := c.managementPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer mgmtConn.Release()

	// for optimisation purposes, try to load connection state and build a map of schemas to load
	// (if we are connected to a remote server running an older CLI,
	// this load may fail, in which case bypass the optimisation)
	connectionStateMap, err := steampipeconfig.LoadConnectionState(ctx, mgmtConn.Conn(), steampipeconfig.WithWaitUntilLoading())
	// NOTE: if we failed to load connection state, this may be because we are connected to an older version of the CLI
	// use legacy (v0.19.x) schema loading code
	if err != nil {
		return c.GetSchemaFromDBLegacy(ctx, mgmtConn)
	}

	// build a ConnectionSchemaMap object to identify the schemas to load
	connectionSchemaMap := steampipeconfig.NewConnectionSchemaMap(ctx, connectionStateMap, c.GetRequiredSessionSearchPath())
	if err != nil {
		return nil, err
	}

	// get the unique schema - we use this to limit the schemas we load from the database
	schemas := maps.Keys(connectionSchemaMap)

	// build a query to retrieve these schemas
	query := c.buildSchemasQuery(schemas...)

	// build schema metadata from query result
	metadata, err := db_common.LoadSchemaMetadata(ctx, mgmtConn.Conn(), query)
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

	// build schema metadata from query result
	return db_common.LoadSchemaMetadata(ctx, conn.Conn(), query)
}

// refreshDbClient terminates the current connection and opens up a new connection to the service.
func (c *DbClient) ResetPools(ctx context.Context) {
	log.Println("[TRACE] db_client.ResetPools start")
	defer log.Println("[TRACE] db_client.ResetPools end")

	if c.userPool != nil {
		c.userPool.Reset()
	}
	if c.managementPool != nil {
		c.managementPool.Reset()
	}
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
