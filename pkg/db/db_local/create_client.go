package db_local

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/constants/runtime"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
)

func getLocalSteampipeConnectionString(opts *CreateDbOptions) (string, error) {
	if opts == nil {
		opts = &CreateDbOptions{}
	}
	utils.LogTime("db.createDbClient start")
	defer utils.LogTime("db.createDbClient end")

	// load the db status
	info, err := GetState()
	if err != nil {
		return "", err
	}
	if info == nil {
		return "", fmt.Errorf("steampipe service is not running")
	}

	// if no database name is passed, use constants.DatabaseUser
	if len(opts.Username) == 0 {
		opts.Username = constants.DatabaseUser
	}
	// if no username name is passed, deduce it from the db status
	if len(opts.DatabaseName) == 0 {
		opts.DatabaseName = info.Database
	}
	// if we still don't have it, fallback to default "postgres"
	if len(opts.DatabaseName) == 0 {
		opts.DatabaseName = "postgres"
	}

	psqlInfoMap := map[string]string{
		// Connect to the database using the first listen address, which is usually localhost
		"host":   info.Listen[0],
		"port":   fmt.Sprintf("%d", info.Port),
		"user":   opts.Username,
		"dbname": opts.DatabaseName,
	}
	psqlInfoMap = utils.MergeMaps(psqlInfoMap, dsnSSLParams())
	log.Println("[TRACE] SQLInfoMap >>>", psqlInfoMap)

	psqlInfo := []string{}
	for k, v := range psqlInfoMap {
		psqlInfo = append(psqlInfo, fmt.Sprintf("%s=%s", k, v))
	}
	log.Println("[TRACE] PSQLInfo >>>", psqlInfo)

	return strings.Join(psqlInfo, " "), nil
}

type CreateDbOptions struct {
	DatabaseName, Username string
}

// createLocalDbClient connects and returns a connection to the given database using
// the provided username
// if the database is not provided (empty), it connects to the default database in the service
// that was created during installation.
// NOTE: no session data callback is used - no sesison data will be present
func createLocalDbClient(ctx context.Context, opts *CreateDbOptions) (*pgx.Conn, error) {
	utils.LogTime("db.createLocalDbClient start")
	defer utils.LogTime("db.createLocalDbClient end")

	psqlInfo, err := getLocalSteampipeConnectionString(opts)
	if err != nil {
		return nil, err
	}

	connConfig, err := pgx.ParseConfig(psqlInfo)
	if err != nil {
		return nil, err
	}

	// set an app name so that we can track database connections from this Steampipe execution
	// this is used to determine whether the database can safely be closed
	connConfig.Config.RuntimeParams = map[string]string{
		"application_name": runtime.PgClientAppName,
	}
	err = db_common.AddRootCertToConfig(&connConfig.Config, getRootCertLocation())
	if err != nil {
		return nil, err
	}

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, err
	}

	if err := db_common.WaitForConnection(ctx, conn); err != nil {
		return nil, err
	}
	return conn, nil
}

// createMaintenanceClient connects to the postgres server using the
// maintenance database and superuser
func createMaintenanceClient(ctx context.Context, port int) (*pgx.Conn, error) {
	utils.LogTime("db_local.createMaintenanceClient start")
	defer utils.LogTime("db_local.createMaintenanceClient end")
	backoff := retry.WithMaxDuration(
		constants.DBConnectionTimeout,
		retry.NewConstant(200*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(ctx, constants.DBConnectionTimeout)
	defer cancel()

	var conn *pgx.Conn

	// create a connection with some retries
	err := retry.Do(ctx, backoff, func(rCtx context.Context) error {
		connStr := fmt.Sprintf("host=localhost port=%d user=%s dbname=postgres sslmode=disable", port, constants.DatabaseSuperUser)
		log.Println("[TRACE] Trying to create maintenance client with: ", connStr)
		dbConnection, err := pgx.Connect(rCtx, connStr)
		if err != nil {
			log.Println("[TRACE] faced error:", err)
			log.Println("[TRACE] retrying:", err)
			return retry.RetryableError(err)
		}
		conn = dbConnection
		return nil
	})

	if err != nil {
		return nil, err
	}

	err = retry.Do(ctx, backoff, func(rCtx context.Context) error {
		if err := db_common.WaitForConnection(rCtx, conn); err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.SQLState() == "57P03" {
				log.Println("[TRACE] looks like a 'cannot_connect_now (57P03):", errors.Unwrap(err))
				// 57P03 is a fatal error that comes up when the database is still starting up
				// let's delay for sometime before trying again
				// using the PingInterval here - can use any other value if required
				time.Sleep(constants.ServicePingInterval)
			}
			return retry.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		conn.Close(ctx)
		return nil, err
	}
	return conn, nil

}
