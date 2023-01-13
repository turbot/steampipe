package db_common

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
)

var ErrServiceInRecovery = errors.New("service is in recovery mode")

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

// WaitForConnection PINGs the DB - retrying after a backoff of constants.ServicePingInterval - but only for constants.DBConnectionTimeout
// returns the error from the database if the dbClient does not respond with after a timeout
func WaitForConnection(ctx context.Context, connection *pgx.Conn) (err error) {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	timeoutCtx, cancel := context.WithTimeout(ctx, constants.DBConnectionTimeout)
	defer func() {
		cancel()
	}()

	retryBackoff := retry.WithMaxDuration(
		constants.DBConnectionTimeout,
		retry.NewConstant(constants.ServicePingInterval),
	)

	retryErr := retry.Do(ctx, retryBackoff, func(ctx context.Context) error {
		log.Println("[TRACE] >>>>>>>>>> Checking IS_IN_RECOVERY")
		InRecoveryMode, err := PgIsInRecovery(ctx, connection)
		if err != nil {
			return retry.RetryableError(err)
		}
		log.Println("[TRACE] >>>>>>>>>> ISINRECOVERY:", InRecoveryMode)
		if InRecoveryMode {
			return retry.RetryableError(ErrServiceInRecovery)
		}
		log.Println("[TRACE] >>>>>>>>>> Pinging")
		pingErr := connection.Ping(timeoutCtx)
		if pingErr != nil {
			return retry.RetryableError(pingErr)
		}
		return nil
	})

	return retryErr
}
