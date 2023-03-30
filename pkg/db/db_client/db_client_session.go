package db_client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
)

func (c *DbClient) AcquireSession(ctx context.Context) (sessionResult *db_common.AcquireSessionResult) {
	sessionResult = &db_common.AcquireSessionResult{}

	defer func() {
		if sessionResult != nil && sessionResult.Session != nil {
			sessionResult.Session.UpdateUsage()

			// fail safe - if there is no database connection, ensure we return an error
			// NOTE: this should not be necessary but an occasional crash is occurring with a nil connection
			if sessionResult.Session.Connection == nil && sessionResult.Error == nil {
				sessionResult.Error = fmt.Errorf("nil database connection being returned from AcquireSession but no error was raised")
			}
		}
	}()

	// save the current value of schema names, then reload the names and check for differences
	schemaNames := c.AllSchemaNames()

	// reload schema names in case they changed based on a connection watcher event
	if err := c.LoadSchemaNames(ctx); err != nil {
		sessionResult.Error = err
		return
	}

	// if the schemas have changes, reset the desired search path
	newSchemaNames := c.AllSchemaNames()
	if !utils.StringSlicesEqual(schemaNames, newSchemaNames) {
		if err := c.updateRequiredSearchPath(ctx); err != nil {
			sessionResult.Error = err
			return
		}
	}

	// get a database connection and query its backend pid
	// note - this will retry if the connection is bad
	databaseConnection, backendPid, err := c.getDatabaseConnectionWithRetries(ctx)
	if err != nil {
		sessionResult.Error = err
		return sessionResult
	}

	c.sessionsMutex.Lock()
	session, found := c.sessions[backendPid]
	if !found {
		session = db_common.NewDBSession(backendPid)
		c.sessions[backendPid] = session
	}
	// we get a new *sql.Conn everytime. USE IT!
	session.Connection = databaseConnection
	sessionResult.Session = session
	c.sessionsMutex.Unlock()

	// make sure that we close the acquired session, in case of error
	defer func() {
		if sessionResult.Error != nil && databaseConnection != nil {
			sessionResult.Session = nil
			databaseConnection.Release()
		}
	}()

	// if the cache is set on the workspace profile
	// then override the default cache setting of the connection
	if viper.IsSet(constants.ArgClientCacheEnabled) {
		if err := db_common.SetCacheEnabled(ctx, viper.GetBool(constants.ArgClientCacheEnabled), databaseConnection.Conn()); err != nil {
			sessionResult.Error = err
			return sessionResult
		}
	}
	if viper.IsSet(constants.ArgCacheTtl) {
		ttl := time.Duration(viper.GetInt(constants.ArgCacheTtl)) * time.Second
		if err := db_common.SetCacheTtl(ctx, ttl, databaseConnection.Conn()); err != nil {
			sessionResult.Error = err
			return sessionResult
		}
	}

	// update required session search path if needed
	err = c.ensureSessionSearchPath(ctx, session)
	if err != nil {
		sessionResult.Error = err
		return sessionResult
	}

	sessionResult.Error = ctx.Err()
	return sessionResult
}

func (c *DbClient) getDatabaseConnectionWithRetries(ctx context.Context) (*pgxpool.Conn, uint32, error) {
	// get a database connection from the pool
	databaseConnection, err := c.pool.Acquire(ctx)
	if err != nil {
		if databaseConnection != nil {
			databaseConnection.Release()
		}
		log.Printf("[TRACE] getDatabaseConnectionWithRetries failed: %s", err.Error())
		return nil, 0, err
	}

	backendPid := databaseConnection.Conn().PgConn().PID()

	return databaseConnection, backendPid, nil
}
