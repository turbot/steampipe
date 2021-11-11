package db_client

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/utils"
)

func (c *DbClient) AcquireSession(ctx context.Context) (_ *sql.Conn, acquireSessionError error) {
	c.sessionInitWaitGroup.Add(1)
	defer c.sessionInitWaitGroup.Done()

	// get a database connection and query its backend pid
	// note - this will retry if the connection is bad
	session, backendPid, err := c.getSessionWithRetries(ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("[WARN] Got Session with PID: %d", backendPid)

	defer func() {
		// make sure that we close the acquired session, in case of error
		if acquireSessionError != nil && session != nil {
			session.Close()
		}
	}()

	if c.ensureSessionFunc == nil {
		return session, nil
	}

	c.sessionMapMutex.Lock()
	sessionStat, isInitialized := c.initializedSessions[backendPid]
	c.sessionMapMutex.Unlock()

	if !isInitialized {
		sessionStat = NewSessionStat()
		log.Printf("[WARN] Session with PID: %d - waiting for init lock", backendPid)
		lockError := c.parallelSessionInitLock.Acquire(ctx, 1)
		if lockError != nil {
			return nil, lockError
		}

		log.Printf("[WARN] Session with PID: %d - got init lock", backendPid)

		err := c.ensureSessionFunc(ctx, session)
		c.parallelSessionInitLock.Release(1)
		if err != nil {
			return nil, err
		}

		log.Printf("[WARN] Session with PID: %d - init DONE", backendPid)

		sessionStat.Initialized = time.Now()
		sessionStat.BackendPid = backendPid
	}

	// update required session search path if needed
	if strings.Join(sessionStat.SearchPath, ",") != strings.Join(c.requiredSessionSearchPath, ",") {
		if err := c.setSessionSearchPathToRequired(ctx, session); err != nil {
			return nil, err
		}
		sessionStat.SearchPath = c.requiredSessionSearchPath
	}

	sessionStat.UpdateUsage()

	// now write back to the map
	c.sessionMapMutex.Lock()
	c.initializedSessions[backendPid] = sessionStat
	c.sessionMapMutex.Unlock()

	log.Printf("[WARN] Session with PID: %d - returning", backendPid)

	return session, nil
}

func (c *DbClient) getSessionWithRetries(ctx context.Context) (*sql.Conn, int64, error) {
	// c.sessionAcquireMutex.Lock()
	// defer c.sessionAcquireMutex.Unlock()

	backoff, err := retry.NewFibonacci(100 * time.Millisecond)
	if err != nil {
		return nil, 0, err
	}

	retries := 0
	var session *sql.Conn
	var backendPid int64
	const getSessionMaxRetries = 10
	err = retry.Do(ctx, retry.WithMaxRetries(getSessionMaxRetries, backoff), func(localCtx context.Context) (e error) {
		if utils.IsContextCancelled(localCtx) {
			return ctx.Err()
		}

		session, err = c.dbClient.Conn(localCtx)
		if err != nil {
			retries++
			return retry.RetryableError(err)
		}
		backendPid, err = GetBackendPid(localCtx, session)
		if err != nil {
			session.Close()
			retries++
			return retry.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		log.Printf("[TRACE] AcquireSession failed after 10 retries: %s", err)
		return nil, 0, err
	}

	if retries > 0 {
		log.Printf("[TRACE] AcquireSession succeeded after %d retries", retries)
	}
	return session, backendPid, nil
}

// get the unique postgres identifier for a database session
func GetBackendPid(ctx context.Context, session *sql.Conn) (int64, error) {
	var pid int64
	rows, err := session.QueryContext(ctx, "select pg_backend_pid()")
	if err != nil {
		return pid, err
	}
	rows.Next()
	rows.Scan(&pid)
	rows.Close()
	return pid, nil
}
