package db_client

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/utils"
)

func (c *DbClient) AcquireSession(ctx context.Context) (_ *db_common.DBSession, acquireSessionError error) {
	c.sessionInitWaitGroup.Add(1)
	defer c.sessionInitWaitGroup.Done()

	// get a database connection and query its backend pid
	// note - this will retry if the connection is bad
	rawConnection, backendPid, err := c.getSessionWithRetries(ctx)
	if err != nil {
		return nil, err
	}

	c.sessionsMutex.Lock()
	session, found := c.sessions[backendPid]
	if !found {
		session = db_common.NewDBSession(backendPid)
		session.Timeline.Add(db_common.DBSessionLifecycleEventCreated)
	}
	// we get a new *sql.Conn everytime. USE IT!
	session.Raw = rawConnection
	c.sessionsMutex.Unlock()

	log.Printf("[TRACE] Got Session with PID: %d", backendPid)

	defer func() {
		// make sure that we close the acquired session, in case of error
		if acquireSessionError != nil && rawConnection != nil {
			rawConnection.Close()
		}
	}()

	if c.ensureSessionFunc == nil {
		return session, nil
	}

	if !session.Initialized {
		log.Printf("[TRACE] Session with PID: %d - waiting for init lock", backendPid)
		session.Timeline.Add(db_common.DBSessionLifecycleEventQueuedForInitialize)

		lockError := c.parallelSessionInitLock.Acquire(ctx, 1)
		if lockError != nil {
			return nil, lockError
		}
		c.sessionInitWaitGroup.Add(1)

		log.Printf("[TRACE] Session with PID: %d - waiting for init start", backendPid)
		session.Timeline.Add(db_common.DBSessionLifecycleEventInitializeStart)
		err := c.ensureSessionFunc(ctx, session)
		session.Timeline.Add(db_common.DBSessionLifecycleEventInitializeFinish)

		session.Initialized = true

		c.sessionInitWaitGroup.Done()
		c.parallelSessionInitLock.Release(1)
		if err != nil {
			return nil, err
		}
		log.Printf("[TRACE] Session with PID: %d - init DONE", backendPid)
	}

	// update required session search path if needed
	if strings.Join(session.SearchPath, ",") != strings.Join(c.requiredSessionSearchPath, ",") {
		if err := c.setSessionSearchPathToRequired(ctx, rawConnection); err != nil {
			return nil, err
		}
		session.SearchPath = c.requiredSessionSearchPath
	}

	session.UpdateUsage()

	// now write back to the map
	c.sessionsMutex.Lock()
	c.sessions[backendPid] = session
	c.sessionsMutex.Unlock()

	log.Printf("[TRACE] Session with PID: %d - returning", backendPid)

	return session, nil
}

func (c *DbClient) getSessionWithRetries(ctx context.Context) (*sql.Conn, int64, error) {
	backoff, err := retry.NewFibonacci(100 * time.Millisecond)
	if err != nil {
		return nil, 0, err
	}

	retries := 0
	var session *sql.Conn
	var backendPid int64
	const getSessionMaxRetries = 10
	err = retry.Do(ctx, retry.WithMaxRetries(getSessionMaxRetries, backoff), func(retryLocalCtx context.Context) (e error) {
		if utils.IsContextCancelled(retryLocalCtx) {
			return ctx.Err()
		}

		session, err = c.dbClient.Conn(retryLocalCtx)
		if err != nil {
			retries++
			return retry.RetryableError(err)
		}
		backendPid, err = db_common.GetBackendPid(retryLocalCtx, session)
		if err != nil {
			session.Close()
			retries++
			return retry.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		log.Printf("[TRACE] getSessionWithRetries failed after 10 retries: %s", err)
		return nil, 0, err
	}

	if retries > 0 {
		log.Printf("[TRACE] getSessionWithRetries succeeded after %d retries", retries)
	}
	return session, backendPid, nil
}
