package db_client

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/utils"
	"golang.org/x/sync/semaphore"
)

// DbClient wraps over `sql.DB` and gives an interface to the database
type DbClient struct {
	connectionString string
	pool             *pgxpool.Pool
	//requiredSessionSearchPath []string

	// concurrency management for db session access
	parallelSessionInitLock *semaphore.Weighted

	// map of database sessions, keyed to the backend_pid in postgres
	// used to update session search path where necessary
	sessions map[uint32]*db_common.DatabaseSession
	// allows locked access to the 'sessions' map
	sessionsMutex *sync.Mutex

	// list of connection schemas
	//foreignSchemaNames []string
	//// list of all local schemas
	//allSchemaNames []string

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

// GetSchemaFromDB requests for all columns of tables backed by steampipe plugins
// and creates golang struct representations from the result
func (c *DbClient) GetSchemaFromDB(ctx context.Context, schemas ...string) (*db_common.SchemaMetadata, error) {
	utils.LogTime("db_client.GetSchemaFromDB start")
	defer utils.LogTime("db_client.GetSchemaFromDB end")
	connection, err := c.pool.Acquire(ctx)
	error_helpers.FailOnError(err)

	query := c.buildSchemasQuery(schemas...)

	tablesResult, err := connection.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	metadata, err := db_common.BuildSchemaMetadata(tablesResult)
	if err != nil {
		return nil, err
	}
	connection.Release()

	return metadata, nil
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
