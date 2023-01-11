package db_common

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
)

// WaitForPool waits for the db to start accepting connections and returns true
// returns false if the dbClient does not start within a stipulated time,
func WaitForPool(ctx context.Context, db *pgxpool.Pool) (err error) {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	pingTimer := time.NewTicker(constants.ServicePingInterval)
	timeoutAt := time.After(constants.DashboardServiceStartTimeout)
	defer pingTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pingTimer.C:
			err = db.Ping(ctx)
			if err == nil {
				return
			}
		case <-timeoutAt:
			return
		}
	}
}

// WaitForConnection waits for the db to start accepting connections and returns true
// returns false if the dbClient does not start within a stipulated time,
func WaitForConnection(ctx context.Context, connection *pgx.Conn) (err error) {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	timeoutCtx, cancel := context.WithTimeout(ctx, constants.DBConnectionTimeout)
	defer func() {
		cancel()
	}()

	return retry.Do(ctx, retry.WithMaxDuration(
		constants.DBConnectionTimeout,
		retry.NewConstant(
			constants.ServicePingInterval,
		),
	), func(ctx context.Context) error {
		err := retry.RetryableError(connection.Ping(timeoutCtx))
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.SQLState() == "57P03" {
				log.Println("[TRACE] faced a 'cannot_connect_now (57P03):", errors.Unwrap(err))
				// 57P03 is a fatal error that comes up when the database is still starting up
				// let's delay for sometime before trying again
				// using the PingInterval here - can use any other value if required
				time.Sleep(constants.ServicePingInterval)
				log.Println("[TRACE] checking again")
			}
			return retry.RetryableError(err)
		}
		return nil
	})
}
