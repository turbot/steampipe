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

var ErrRecoveryMode = errors.New("service is in recovery mode")

// WaitForPool waits for the db to start accepting connections and returns true
// returns false if the dbClient does not start within a stipulated time,
func WaitForPool(ctx context.Context, db *pgxpool.Pool) (err error) {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	pingTimer := time.NewTicker(constants.ServicePingInterval)
	timeoutAt := time.After(constants.DBConnectionTimeout)
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

// WaitForConnectionPing PINGs the DB - retrying after a backoff of constants.ServicePingInterval - but only for constants.DBConnectionTimeout
// returns the error from the database if the dbClient does not respond successfully after a timeout
func WaitForConnectionPing(ctx context.Context, connection *pgx.Conn) (err error) {
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
		log.Println("[TRACE] Pinging")
		pingErr := connection.Ping(timeoutCtx)
		if pingErr != nil {
			log.Println("[TRACE] Pinging failed -> trying again")
			return retry.RetryableError(pingErr)
		}
		return nil
	})

	return retryErr
}

// WaitForRecovery returns an error (ErrRecoveryMode) if the service stays in recovery
// mode for more than constants.DBRecoveryWaitTimeout
func WaitForRecovery(ctx context.Context, connection *pgx.Conn) (err error) {
	utils.LogTime("db_common.WaitForRecovery start")
	defer utils.LogTime("db_common.WaitForRecovery end")

	timeoutCtx, cancel := context.WithTimeout(ctx, constants.DBRecoveryWaitTimeout)
	defer func() {
		cancel()
	}()

	retryBackoff := retry.WithMaxDuration(
		constants.DBRecoveryWaitTimeout,
		retry.NewConstant(constants.ServicePingInterval),
	)

	retryErr := retry.Do(timeoutCtx, retryBackoff, func(ctx context.Context) error {
		log.Println("[TRACE] checking for recovery mode")
		row := connection.QueryRow(ctx, "select pg_is_in_recovery();")
		var isInRecovery bool
		if scanErr := row.Scan(&isInRecovery); scanErr != nil {
			return retry.RetryableError(scanErr)
		}
		if isInRecovery {
			return retry.RetryableError(ErrRecoveryMode)
		}
		return nil
	})

	return retryErr
}
