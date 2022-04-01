package db_client

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4/stdlib"

	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/utils"
)

func (c *DbClient) AcquireSession(ctx context.Context) (sessionResult *db_common.AcquireSessionResult) {
	sessionResult = &db_common.AcquireSessionResult{}
	c.sessionInitWaitGroup.Add(1)
	defer c.sessionInitWaitGroup.Done()

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

	// reload foreign schema names in case they changed based on a connection watcher event
	if err := c.LoadForeignSchemaNames(ctx); err != nil {
		sessionResult.Error = err
		return
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
		session.LifeCycle.Add("created")
	}
	// we get a new *sql.Conn everytime. USE IT!
	session.Connection = databaseConnection
	sessionResult.Session = session
	c.sessionsMutex.Unlock()

	// make sure that we close the acquired session, in case of error
	defer func() {
		if sessionResult.Error != nil && databaseConnection != nil {
			databaseConnection.Close()
		}
	}()

	// if there is no ensure session function, we are done
	if c.ensureSessionFunc == nil {
		return sessionResult
	}

	if !session.Initialized {
		session.LifeCycle.Add("queued_for_init")

		err := c.parallelSessionInitLock.Acquire(ctx, 1)
		if err != nil {
			sessionResult.Error = err
			return sessionResult
		}
		c.sessionInitWaitGroup.Add(1)

		session.LifeCycle.Add("init_start")
		err, warnings := c.ensureSessionFunc(ctx, session)
		session.LifeCycle.Add("init_finish")
		sessionResult.Warnings = warnings
		c.sessionInitWaitGroup.Done()
		c.parallelSessionInitLock.Release(1)
		if err != nil {
			sessionResult.Error = err
			return sessionResult
		}

		// if there is no error, mark session as initialized
		session.Initialized = true
	}

	// update required session search path if needed
	err = c.ensureSessionSearchPath(ctx, session)
	if err != nil {
		sessionResult.Error = err
		return sessionResult
	}

	// now write back to the map
	c.sessionsMutex.Lock()
	c.sessions[backendPid] = session
	c.sessionsMutex.Unlock()

	return sessionResult
}

func (c *DbClient) getDatabaseConnectionWithRetries(ctx context.Context) (*sql.Conn, uint32, error) {
	backoff, err := retry.NewFibonacci(100 * time.Millisecond)
	if err != nil {
		return nil, 0, err
	}

	var databaseConnection *sql.Conn
	var backendPid uint32

	retries := 0
	const getSessionMaxRetries = 10
	err = retry.Do(ctx, retry.WithMaxRetries(getSessionMaxRetries, backoff), func(retryLocalCtx context.Context) (e error) {
		if utils.IsContextCancelled(retryLocalCtx) {
			return retryLocalCtx.Err()
		}
		// get a database connection from the pool
		databaseConnection, err = c.dbClient.Conn(retryLocalCtx)
		if err != nil {
			if databaseConnection != nil {
				databaseConnection.Close()
			}
			retries++
			return retry.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		log.Printf("[TRACE] getDatabaseConnectionWithRetries failed after %d retries: %s", retries, err)
		return nil, 0, err
	}

	if retries > 0 {
		log.Printf("[TRACE] getDatabaseConnectionWithRetries succeeded after %d retries", retries)
	}

	databaseConnection.Raw(func(driverConn interface{}) error {
		backendPid = driverConn.(*stdlib.Conn).Conn().PgConn().PID()
		return nil
	})

	return databaseConnection, uint32(backendPid), nil
}
