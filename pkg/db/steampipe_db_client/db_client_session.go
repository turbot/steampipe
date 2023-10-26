package steampipe_db_client

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/turbot/steampipe/pkg/db/steampipe_db_common"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
)

func (c *SteampipeDbClient) AcquireManagementConnection(ctx context.Context) (*sql.Conn, error) {
	return c.ManagementPool.Conn(ctx)
}

func (c *SteampipeDbClient) AcquireSession(ctx context.Context) (sessionResult *steampipe_db_common.AcquireSessionResult) {
	sessionResult = &steampipe_db_common.AcquireSessionResult{}

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
	databaseConnection, err := c.UserPool.Conn(ctx)
	if err != nil {
		sessionResult.Error = err
		return sessionResult
	}

	// backendPid := databaseConnection.Conn().PgConn().PID()
	// c.sessionsMutex.Lock()
	// session, found := c.sessions[backendPid]
	// if !found {
	// 	session = db_common.NewDBSession(backendPid)
	// 	c.sessions[backendPid] = session
	// }
	// c.sessionsMutex.Unlock()

	// we get a new *sql.Conn everytime. USE IT!
	session := steampipe_db_common.NewDBSession(0)
	session.Connection = databaseConnection
	sessionResult.Session = session

	// make sure that we close the acquired session, in case of error
	defer func() {
		if sessionResult.Error != nil && databaseConnection != nil {
			sessionResult.Session = nil
			databaseConnection.Close()
		}
	}()

	// if this is connected to a local service (localhost) and if the server cache
	// is disabled, override the client setting to always disable
	//
	// this is a temporary workaround to make sure
	// that we turn off caching for plugins compiled with SDK pre-V5
	if c.isLocalService && !viper.GetBool(constants.ArgServiceCacheEnabled) {
		if err := steampipe_db_common.SetCacheEnabled(ctx, false, databaseConnection); err != nil {
			sessionResult.Error = err
			return sessionResult
		}
	} else {
		if viper.IsSet(constants.ArgClientCacheEnabled) {
			if err := steampipe_db_common.SetCacheEnabled(ctx, viper.GetBool(constants.ArgClientCacheEnabled), databaseConnection); err != nil {
				sessionResult.Error = err
				return sessionResult
			}
		}
	}

	if viper.IsSet(constants.ArgCacheTtl) {
		ttl := time.Duration(viper.GetInt(constants.ArgCacheTtl)) * time.Second
		if err := steampipe_db_common.SetCacheTtl(ctx, ttl, databaseConnection); err != nil {
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
