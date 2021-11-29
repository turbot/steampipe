package db_client

import (
	"context"
	"database/sql"
	"sync"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/steampipeconfig"
	"golang.org/x/sync/semaphore"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
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
	sessions map[int64]*db_common.DatabaseSession
	// allows locked access to the 'sessions' map
	sessionsMutex *sync.Mutex

	// list of connection schemas
	schemas    []string
	searchPath []string
}

func NewDbClient(connectionString string) (*DbClient, error) {
	utils.LogTime("db_client.NewDbClient start")
	defer utils.LogTime("db_client.NewDbClient end")
	db, err := establishConnection(connectionString)
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
		sessions:                make(map[int64]*db_common.DatabaseSession),
		sessionsMutex:           &sync.Mutex{},
	}
	client.connectionString = connectionString

	// TODO populate schemas
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

// Schemas implements Client
func (c *DbClient) Schemas() []string {
	return c.schemas
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
		sessionResult.Session.Close()
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
	db, err := establishConnection(c.connectionString)
	if err != nil {
		return err
	}
	c.dbClient = db

	return nil
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
