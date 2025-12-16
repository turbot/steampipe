package db_client

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

func (c *DbClient) AcquireManagementConnection(ctx context.Context) (*pgxpool.Conn, error) {
	return c.managementPool.Acquire(ctx)
}

func (c *DbClient) AcquireSession(ctx context.Context) (sessionResult *db_common.AcquireSessionResult) {
	sessionResult = &db_common.AcquireSessionResult{}

	defer func() {
		if sessionResult != nil && sessionResult.Session != nil {
			// fail safe - if there is no database connection, ensure we return an error
			// NOTE: this should not be necessary but an occasional crash is occurring with a nil connection
			if sessionResult.Session.Connection == nil && sessionResult.Error == nil {
				sessionResult.Error = fmt.Errorf("nil database connection being returned from AcquireSession but no error was raised")
			}
		}
	}()

	// get a database connection and query its backend pid
	// note - this will retry if the connection is bad
	databaseConnection, err := c.userPool.Acquire(ctx)
	if err != nil {
		sessionResult.Error = err
		return sessionResult
	}
	backendPid := databaseConnection.Conn().PgConn().PID()

	c.lockSessions()
	// Check if client has been closed (sessions set to nil)
	if c.sessions == nil {
		c.sessionsUnlock()
		sessionResult.Error = fmt.Errorf("client has been closed")
		return sessionResult
	}
	session, found := c.sessions[backendPid]
	if !found {
		session = db_common.NewDBSession(backendPid)
		c.sessions[backendPid] = session
	}
	// we get a new *sql.Conn everytime. USE IT!
	session.Connection = databaseConnection
	sessionResult.Session = session
	c.sessionsUnlock()

	// make sure that we close the acquired session, in case of error
	defer func() {
		if sessionResult.Error != nil && databaseConnection != nil {
			sessionResult.Session = nil
			databaseConnection.Release()
		}
	}()

	// if this is connected to a local service (localhost) and if the server cache
	// is disabled, override the client setting to always disable
	//
	// this is a temporary workaround to make sure
	// that we turn off caching for plugins compiled with SDK pre-V5
	if c.isLocalService && !viper.GetBool(constants.ArgServiceCacheEnabled) {
		if err := db_common.SetCacheEnabled(ctx, false, databaseConnection.Conn()); err != nil {
			sessionResult.Error = err
			return sessionResult
		}
	} else {
		if viper.IsSet(constants.ArgClientCacheEnabled) {
			if err := db_common.SetCacheEnabled(ctx, viper.GetBool(constants.ArgClientCacheEnabled), databaseConnection.Conn()); err != nil {
				sessionResult.Error = err
				return sessionResult
			}
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
