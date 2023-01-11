package db_common

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
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

	pingTimer := time.NewTicker(constants.ServicePingInterval)
	timeoutCtx, cancel := context.WithTimeout(ctx, constants.DBConnectionTimeout)
	defer func() {
		cancel()
		// prevent the timer from leaking
		pingTimer.Stop()
	}()

	for {
		select {
		case <-timeoutCtx.Done():
			return errors.Wrap(ctx.Err(), "WaitForConnection timed out")
		case <-pingTimer.C:
			err = connection.Ping(timeoutCtx)
			if err == nil {
				return
			}
		}
	}
}
