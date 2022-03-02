package db_local

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_client"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

// LocalDbClient wraps over DbClient
type LocalDbClient struct {
	client        *db_client.DbClient
	invoker       constants.Invoker
	connectionMap *steampipeconfig.ConnectionDataMap
}

// GetLocalClient starts service if needed and creates a new LocalDbClient
func GetLocalClient(ctx context.Context, invoker constants.Invoker) (db_common.Client, error) {
	utils.LogTime("db.GetLocalClient start")
	defer utils.LogTime("db.GetLocalClient end")

	// start db if necessary
	if err := EnsureDBInstalled(ctx); err != nil {
		return nil, err
	}

	startResult := StartServices(ctx, constants.DatabaseDefaultPort, ListenTypeLocal, invoker)
	if startResult.Error != nil {
		return nil, startResult.Error
	}

	client, err := NewLocalClient(ctx, invoker)
	if err != nil {
		ShutdownService(ctx, invoker)
	}
	return client, err
}

// NewLocalClient verifies that the local database instance is running and returns a Client to interact with it
func NewLocalClient(ctx context.Context, invoker constants.Invoker) (*LocalDbClient, error) {
	utils.LogTime("db.NewLocalClient start")
	defer utils.LogTime("db.NewLocalClient end")

	connString, err := getLocalSteampipeConnectionString()
	if err != nil {
		return nil, err
	}
	dbClient, err := db_client.NewDbClient(ctx, connString)
	if err != nil {
		log.Printf("[WARN] error getting local client %s", err.Error())
		return nil, err
	}

	c := &LocalDbClient{client: dbClient, invoker: invoker}
	log.Printf("[TRACE] created local client %p", c)
	return c, nil
}

// Close implements Client
// close the connection to the database and shuts down the backend
func (c *LocalDbClient) Close(ctx context.Context) error {
	log.Printf("[TRACE] close local client %p", c)
	if c.client != nil {
		log.Printf("[TRACE] local client not NIL")
		if err := c.client.Close(ctx); err != nil {
			return err
		}
		log.Printf("[TRACE] local client close complete")
	}
	log.Printf("[TRACE] shutdown local service %v", c.invoker)
	ShutdownService(ctx, c.invoker)
	return nil
}

// SetEnsureSessionDataFunc implements Client
func (c *LocalDbClient) SetEnsureSessionDataFunc(f db_common.EnsureSessionStateCallback) {
	c.client.SetEnsureSessionDataFunc(f)
}

// ForeignSchemas implements Client
func (c *LocalDbClient) ForeignSchemas() []string {
	return c.client.ForeignSchemas()
}

func (c *LocalDbClient) ConnectionMap() *steampipeconfig.ConnectionDataMap {
	return c.connectionMap
}

func (c *LocalDbClient) RefreshSessions(ctx context.Context) *db_common.AcquireSessionResult {
	return c.client.RefreshSessions(ctx)
}

func (c *LocalDbClient) AcquireSession(ctx context.Context) *db_common.AcquireSessionResult {
	return c.client.AcquireSession(ctx)
}

// ExecuteSync implements Client
func (c *LocalDbClient) ExecuteSync(ctx context.Context, query string) (*queryresult.SyncQueryResult, error) {
	return c.client.ExecuteSync(ctx, query)
}

// ExecuteSyncInSession implements Client
func (c *LocalDbClient) ExecuteSyncInSession(ctx context.Context, session *db_common.DatabaseSession, query string) (*queryresult.SyncQueryResult, error) {
	return c.client.ExecuteSyncInSession(ctx, session, query)
}

// ExecuteInSession implements Client
func (c *LocalDbClient) ExecuteInSession(ctx context.Context, session *db_common.DatabaseSession, query string, onComplete func()) (res *queryresult.Result, err error) {
	return c.client.ExecuteInSession(ctx, session, query, onComplete)
}

// Execute implements Client
func (c *LocalDbClient) Execute(ctx context.Context, query string) (res *queryresult.Result, err error) {
	return c.client.Execute(ctx, query)
}

// CacheOn implements Client
func (c *LocalDbClient) CacheOn(ctx context.Context) error {
	return c.client.CacheOn(ctx)
}

// CacheOff implements Client
func (c *LocalDbClient) CacheOff(ctx context.Context) error {
	return c.client.CacheOff(ctx)
}

// CacheClear implements Client
func (c *LocalDbClient) CacheClear(ctx context.Context) error {
	return c.client.CacheClear(ctx)
}

// GetCurrentSearchPath implements Client
func (c *LocalDbClient) GetCurrentSearchPath(ctx context.Context) ([]string, error) {
	return c.client.GetCurrentSearchPath(ctx)
}

// SetSessionSearchPath implements Client
func (c *LocalDbClient) SetSessionSearchPath(ctx context.Context, currentUserPath ...string) error {
	return c.client.SetSessionSearchPath(ctx, currentUserPath...)
}

func (c *LocalDbClient) ContructSearchPath(ctx context.Context, requiredSearchPath []string, searchPathPrefix []string, currentSearchPath []string) ([]string, error) {
	return c.client.ContructSearchPath(ctx, requiredSearchPath, searchPathPrefix, currentSearchPath)
}

// GetSchemaFromDB for LocalDBClient optimises the schema extraction by extracting schema
// information for connections backed by distinct plugins and then fanning back out.
func (c *LocalDbClient) GetSchemaFromDB(ctx context.Context) (*schema.Metadata, error) {
	// build a ConnectionSchemaMap object to identify the schemas to load
	// (pass nil for connection state - this forces NewConnectionSchemaMap to load it)
	connectionSchemaMap, err := steampipeconfig.NewConnectionSchemaMap()
	if err != nil {
		return nil, err
	}
	// get the unique schema - we use this to limit the schemas we load from the database
	schemas := connectionSchemaMap.UniqueSchemas()
	query := c.buildSchemasQuery(schemas)

	acquireSessionResult := c.AcquireSession(ctx)
	if acquireSessionResult.Error != nil {
		acquireSessionResult.Session.Close(false)
		return nil, err
	}

	tablesResult, err := acquireSessionResult.Session.Connection.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	metadata, err := db_common.BuildSchemaMetadata(tablesResult)
	if err != nil {
		acquireSessionResult.Session.Close(false)
		return nil, err
	}
	acquireSessionResult.Session.Close(false)

	c.populateSchemaMetadata(metadata, connectionSchemaMap)

	searchPath, err := c.GetCurrentSearchPath(ctx)
	if err != nil {
		return nil, err
	}
	metadata.SearchPath = searchPath

	return metadata, nil
}

// update schemaMetadata to add in all other schemas which have the same schemas as those we have loaded
// NOTE: this mutates 	schemaMetadata
func (c *LocalDbClient) populateSchemaMetadata(schemaMetadata *schema.Metadata, connectionSchemaMap steampipeconfig.ConnectionSchemaMap) {
	// we now need to add in all other schemas which have the same schemas as those we have loaded
	for loadedSchema, otherSchemas := range connectionSchemaMap {
		// all 'otherSchema's have the same schema as loadedSchema
		exemplarSchema, ok := schemaMetadata.Schemas[loadedSchema]
		if !ok {
			// should can happen in the case of a dynamic plugin with no tables - use empty schema
			exemplarSchema = make(map[string]schema.TableSchema)
		}

		for _, s := range otherSchemas {
			schemaMetadata.Schemas[s] = exemplarSchema
		}
	}
}

func (c *LocalDbClient) buildSchemasQuery(schemas []string) string {
	for idx, s := range schemas {
		schemas[idx] = fmt.Sprintf("'%s'", s)
	}
	schemaClause := strings.Join(schemas, ",")
	query := fmt.Sprintf(`
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
	cols.table_schema in (%s)
	OR
    LEFT(cols.table_schema,8) = 'pg_temp_'

`, schemaClause)
	return query
}

func (c *LocalDbClient) LoadForeignSchemaNames(ctx context.Context) error {
	return c.client.LoadForeignSchemaNames(ctx)
}

// local only functions

func (c *LocalDbClient) RefreshConnectionAndSearchPaths(ctx context.Context) *steampipeconfig.RefreshConnectionResult {
	// NOTE: disable any status updates - we do not want 'loading' output from any queries
	ctx = statushooks.DisableStatusHooks(ctx)

	res := c.refreshConnections(ctx)
	if res.Error != nil {
		return res
	}
	if err := refreshFunctions(ctx); err != nil {
		res.Error = err
		return res
	}

	// load the connection state and cache it!
	connectionMap, err := steampipeconfig.GetConnectionState(c.ForeignSchemas())
	if err != nil {
		res.Error = err
		return res
	}
	c.connectionMap = &connectionMap
	// set user search path first - client may fall back to using it
	currentSearchPath, err := c.setUserSearchPath(ctx)
	if err != nil {
		res.Error = err
		return res
	}

	if err := c.SetSessionSearchPath(ctx, currentSearchPath...); err != nil {
		res.Error = err
		return res
	}

	return res
}

// SetUserSearchPath sets the search path for the all steampipe users of the db service
// do this wy finding all users assigned to the role steampipe_users and set their search path
func (c *LocalDbClient) setUserSearchPath(ctx context.Context) ([]string, error) {
	var searchPath []string

	// is there a user search path in the config?
	// check ConfigKeyDatabaseSearchPath config (this is the value specified in the database config)
	if viper.IsSet(constants.ConfigKeyDatabaseSearchPath) {
		searchPath = viper.GetStringSlice(constants.ConfigKeyDatabaseSearchPath)
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.FunctionSchema)
	} else {
		// no config set - set user search path to default
		searchPath = c.getDefaultSearchPath()
	}

	// escape the schema names
	escapedSearchPath := db_common.PgEscapeSearchPath(searchPath)

	log.Println("[TRACE] setting user search path to", searchPath)

	// get all roles which are a member of steampipe_users
	query := fmt.Sprintf(`select usename from pg_user where pg_has_role(usename, '%s', 'member')`, constants.DatabaseUsersRole)
	res, err := c.ExecuteSync(context.Background(), query)
	if err != nil {
		return nil, err
	}

	// set the search path for all these roles
	var queries []string
	for _, row := range res.Rows {
		rowResult := row.(*queryresult.RowResult)
		user := string(rowResult.Data[0].(string))
		if user == "root" {
			continue
		}
		queries = append(queries, fmt.Sprintf(
			"alter user %s set search_path to %s;",
			db_common.PgEscapeName(user),
			strings.Join(escapedSearchPath, ","),
		))
	}
	query = strings.Join(queries, "\n")
	log.Printf("[TRACE] user search path sql: %s", query)
	_, err = executeSqlAsRoot(ctx, query)
	if err != nil {
		return nil, err
	}
	return searchPath, nil
}

// build default search path from the connection schemas, bookended with public and internal
func (c *LocalDbClient) getDefaultSearchPath() []string {
	searchPath := c.ForeignSchemas()
	sort.Strings(searchPath)
	// add the 'public' schema as the first schema in the search_path. This makes it
	// easier for users to build and work with their own tables, and since it's normally
	// empty, doesn't make using steampipe tables any more difficult.
	searchPath = append([]string{"public"}, searchPath...)
	// add 'internal' schema as last schema in the search path
	searchPath = append(searchPath, constants.FunctionSchema)

	return searchPath
}
