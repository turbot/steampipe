package db_common

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
)

var ErrServiceInRecoveryMode = errors.New("service is in recovery mode")

type waitConfig struct {
	retryInterval time.Duration
	timeout       time.Duration
}

type WaitOption func(w *waitConfig)

func WithRetryInterval(d time.Duration) WaitOption {
	return func(w *waitConfig) {
		w.retryInterval = d
	}
}
func WithTimeout(d time.Duration) WaitOption {
	return func(w *waitConfig) {
		w.timeout = d
	}
}

// WaitForPool waits for the db to start accepting connections and returns true
// returns false if the dbClient does not start within a stipulated time,
func WaitForPool(ctx context.Context, db *sql.DB, waitOptions ...WaitOption) (err error) {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	var connection *sql.Conn
	backoff := retry.WithMaxDuration(
		constants.DBStartTimeout,
		retry.NewConstant(constants.DBConnectionRetryBackoff),
	)

	// if we are here, we are more or less sure that the database is up and running
	// but we need to wait for it to accept connections
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		c, err := db.Conn(ctx)
		if err != nil {
			return retry.RetryableError(err)
		}
		connection = c
		return nil
	})

	defer connection.Close()
	return WaitForConnectionPing(ctx, connection, waitOptions...)
}

// WaitForConnectionPing PINGs the DB - retrying after a backoff of constants.ServicePingInterval - but only for constants.DBConnectionTimeout
// returns the error from the database if the dbClient does not respond successfully after a timeout
func WaitForConnectionPing(ctx context.Context, connection *sql.Conn, waitOptions ...WaitOption) (err error) {
	utils.LogTime("db_common.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	config := &waitConfig{
		retryInterval: constants.ServicePingInterval,
		timeout:       constants.DBStartTimeout,
	}

	for _, o := range waitOptions {
		o(config)
	}

	retryBackoff := retry.WithMaxDuration(
		config.timeout,
		retry.NewConstant(config.retryInterval),
	)

	retryErr := retry.Do(ctx, retryBackoff, func(ctx context.Context) error {
		log.Println("[TRACE] Pinging")
		pingErr := connection.PingContext(ctx)
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
func WaitForRecovery(ctx context.Context, connection *sql.Conn, waitOptions ...WaitOption) (err error) {
	utils.LogTime("db_common.WaitForRecovery start")
	defer utils.LogTime("db_common.WaitForRecovery end")

	config := &waitConfig{
		retryInterval: constants.ServicePingInterval,
		timeout:       time.Duration(0),
	}

	for _, o := range waitOptions {
		o(config)
	}

	var retryBackoff retry.Backoff
	if config.timeout == 0 {
		retryBackoff = retry.NewConstant(config.retryInterval)
	} else {
		retryBackoff = retry.WithMaxDuration(
			config.timeout,
			retry.NewConstant(config.retryInterval),
		)
	}

	// this is to make sure that we set the
	// "recovering" status only once, even if it's
	// called from inside the retry loop
	recoveryStatusUpdateOnce := &sync.Once{}

	retryErr := retry.Do(ctx, retryBackoff, func(ctx context.Context) error {
		log.Println("[TRACE] checking for recovery mode")
		row := connection.QueryRowContext(ctx, "select pg_is_in_recovery();")
		var isInRecovery bool
		if scanErr := row.Scan(&isInRecovery); scanErr != nil {
			if error_helpers.IsContextCancelledError(scanErr) {
				return scanErr
			}
			log.Println("[ERROR] checking for recover mode", scanErr)
			return retry.RetryableError(scanErr)
		}
		if isInRecovery {
			log.Println("[TRACE] service is in recovery")

			recoveryStatusUpdateOnce.Do(func() {
				statushooks.SetStatus(ctx, "Database is recovering. This may take some time.")
			})

			return retry.RetryableError(ErrServiceInRecoveryMode)
		}
		return nil
	})

	return retryErr
}
