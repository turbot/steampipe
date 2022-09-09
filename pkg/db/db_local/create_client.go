package db_local

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
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

	// Connect to the database using the first listen address, which is usually localhost
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
		info.Listen[0],
		info.Port,
		opts.Username,
		opts.DatabaseName,
		sslMode())

	return psqlInfo, nil
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

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, err
	}

	if err := db_common.WaitForConnection(ctx, conn); err != nil {
		return nil, err
	}
	return conn, nil
}
