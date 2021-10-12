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
func GetLocalClient(invoker constants.Invoker) (db_common.Client, error) {
	// start db if necessary
	err := EnsureDbAndStartService(invoker)
	if err != nil {
		return nil, err
	}

	client, err := NewLocalClient(invoker)
	if err != nil {
		ShutdownService(invoker)
	}
	// NOTE:  client shutdown will shutdown service (if invoker matches)
	return client, nil
}

// NewLocalClient ensures that the database instance is running
// and returns a `Client` to interact with it
func NewLocalClient(invoker constants.Invoker) (*LocalDbClient, error) {
	utils.LogTime("db.NewLocalClient start")
	defer utils.LogTime("db.NewLocalClient end")

	connString, err := getLocalSteampipeConnectionString()
	if err != nil {
		return nil, err
	}
	dbClient, err := db_client.NewDbClient(connString)
	if err != nil {
		return nil, err
	}

	c := &LocalDbClient{client: dbClient, invoker: invoker}
	log.Printf("[TRACE] created local client %p", c)
	return c, nil
}

// Close implements Client
// close the connection to the database and shuts down the backend
func (c *LocalDbClient) Close() error {
	log.Printf("[TRACE] close local client %p", c)
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return err
		}
	}
	ShutdownService(c.invoker)
	return nil
}

// EnsureSessionState implements Client
func (c *LocalDbClient) SetEnsureSessionDataFunc(f db_common.EnsureSessionStateCallback) {
	c.client.SetEnsureSessionDataFunc(f)
}

// SchemaMetadata implements Client
func (c *LocalDbClient) SchemaMetadata() *schema.Metadata {
	return c.client.SchemaMetadata()
}

func (c *LocalDbClient) ConnectionMap() *steampipeconfig.ConnectionDataMap {
	return c.connectionMap
}

// LoadSchema  implements Client
func (c *LocalDbClient) LoadSchema() {
	c.client.LoadSchema()
}

func (c *LocalDbClient) RefreshSessions(ctx context.Context) error {
	return c.client.RefreshSessions(ctx)
}

// ExecuteSync implements Client
func (c *LocalDbClient) ExecuteSync(ctx context.Context, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error) {
	return c.client.ExecuteSync(ctx, query, disableSpinner)
}

// Execute implements Client
func (c *LocalDbClient) Execute(ctx context.Context, query string, disableSpinner bool) (res *queryresult.Result, err error) {
	return c.client.Execute(ctx, query, disableSpinner)
}

// CacheOn implements Client
func (c *LocalDbClient) CacheOn() error {
	return c.client.CacheOn()
}

// CacheOff implements Client
func (c *LocalDbClient) CacheOff() error {
	return c.client.CacheOff()
}

// CacheClear implements Client
func (c *LocalDbClient) CacheClear() error {
	return c.client.CacheClear()
}

// GetCurrentSearchPath implements Client
func (c *LocalDbClient) GetCurrentSearchPath() ([]string, error) {
	// NOTE: create a new client to do this, so we respond to any recent changes in user search path
	// (as the user search path may have changed  after creating client 'c', e.g. if connections have changed)
	newClient, err := NewLocalClient(constants.InvokerService)
	if err != nil {
		return nil, err
	}
	defer newClient.Close()
	return newClient.client.GetCurrentSearchPath()
}

// SetSessionSearchPath implements Client
func (c *LocalDbClient) SetSessionSearchPath(currentSearchPath ...string) error {
	return c.client.SetSessionSearchPath(currentSearchPath...)
}

// local only functions

func (c *LocalDbClient) RefreshConnectionAndSearchPaths() *steampipeconfig.RefreshConnectionResult {
	res := c.refreshConnections()
	if res.Error != nil {
		return res
	}
	if err := refreshFunctions(); err != nil {
		res.Error = err
		return res
	}

	// load the connection state and cache it!
	connectionMap, err := steampipeconfig.GetConnectionState(c.SchemaMetadata().GetSchemas())
	if err != nil {
		res.Error = err
		return res
	}
	c.connectionMap = &connectionMap
	// set user search path first - client may fall back to using it
	if err := c.setUserSearchPath(); err != nil {
		res.Error = err
		return res
	}

	// get current search path, creating a new client to ensure we pick up recent changes
	currentSearchPath, err := c.GetCurrentSearchPath()
	if err != nil {
		res.Error = err
		return res
	}
	if err := c.SetSessionSearchPath(currentSearchPath...); err != nil {
		res.Error = err
		return res
	}

	return res
}

// SetUserSearchPath sets the search path for the all steampipe users of the db service
// do this wy finding all users assigned to the role steampipe_users and set their search path
func (c *LocalDbClient) setUserSearchPath() error {
	log.Println("[TRACE] SetUserSearchPath")
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
	searchPath = db_common.PgEscapeSearchPath(searchPath)

	log.Println("[TRACE] setting user search path to", searchPath)

	// get all roles which are a member of steampipe_users
	query := fmt.Sprintf(`select usename from pg_user where pg_has_role(usename, '%s', 'member')`, constants.DatabaseUsersRole)
	res, err := c.ExecuteSync(context.Background(), query, true)
	if err != nil {
		return err
	}

	// set the search path for all these roles
	var queries []string
	for _, row := range res.Rows {
		rowResult := row.(*queryresult.RowResult)
		user := string(rowResult.Data[0].([]uint8))
		if user == "root" {
			continue
		}
		queries = append(queries, fmt.Sprintf(
			"alter user %s set search_path to %s;",
			user,
			strings.Join(searchPath, ","),
		))
	}
	query = strings.Join(queries, "\n")
	log.Printf("[TRACE] user search path sql: %s", query)
	_, err = executeSqlAsRoot(query)
	if err != nil {
		return err
	}
	return nil
}

// build default search path from the connection schemas, bookended with public and internal
func (c *LocalDbClient) getDefaultSearchPath() []string {
	searchPath := c.SchemaMetadata().GetSchemas()
	sort.Strings(searchPath)
	// add the 'public' schema as the first schema in the search_path. This makes it
	// easier for users to build and work with their own tables, and since it's normally
	// empty, doesn't make using steampipe tables any more difficult.
	searchPath = append([]string{"public"}, searchPath...)
	// add 'internal' schema as last schema in the search path
	searchPath = append(searchPath, constants.FunctionSchema)

	return searchPath
}
