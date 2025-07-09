package db_common

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
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

func WaitForConnection(ctx context.Context, connStr string, options ...WaitOption) (conn *pgx.Conn, err error) {
	utils.LogTime("db_common.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	config := &waitConfig{
		retryInterval: constants.DBConnectionRetryBackoff,
		timeout:       constants.DBStartTimeout,
	}

	for _, o := range options {
		o(config)
	}

	backoff := retry.WithMaxDuration(
		config.timeout,
		retry.NewConstant(config.retryInterval),
	)

	// create a connection to the service.
	// Retry after a backoff, but only upto a maximum duration.
	err = retry.Do(ctx, backoff, func(rCtx context.Context) error {
		log.Println("[TRACE] Trying to create client with: ", connStr)
		dbConnection, err := pgx.Connect(rCtx, connStr)
		if err != nil {
			log.Println("[TRACE] could not connect:", err)
			return retry.RetryableError(err)
		}
		log.Println("[TRACE] connected to database")
		conn = dbConnection
		return nil
	})

	return conn, err
}

// WaitForPool waits for the db to start accepting connections and returns true
// returns false if the dbClient does not start within a stipulated time,
func WaitForPool(ctx context.Context, db *pgxpool.Pool, waitOptions ...WaitOption) (err error) {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	connection, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer connection.Release()
	return WaitForConnectionPing(ctx, connection.Conn(), waitOptions...)
}

// WaitForConnectionPing PINGs the DB - retrying after a backoff of constants.ServicePingInterval - but only for constants.DBConnectionTimeout
// returns the error from the database if the dbClient does not respond successfully after a timeout
func WaitForConnectionPing(ctx context.Context, connection *pgx.Conn, waitOptions ...WaitOption) (err error) {
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
		pingErr := connection.Ping(ctx)
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
func WaitForRecovery(ctx context.Context, connection *pgx.Conn, waitOptions ...WaitOption) (err error) {
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
		row := connection.QueryRow(ctx, "select pg_is_in_recovery();")
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
