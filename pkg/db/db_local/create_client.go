package db_local

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/turbot/steampipe/pkg/db/db_client"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
)

func getLocalSteampipeConnectionString(opts *CreateDbOptions) (string, error) {
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

	// if no database name is passed, deduce it from the db status
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
func createLocalDbClient(ctx context.Context, opts *CreateDbOptions) (*pgxpool.Pool, error) {
	utils.LogTime("db.createLocalDbClient start")
	defer utils.LogTime("db.createLocalDbClient end")

	psqlInfo, err := getLocalSteampipeConnectionString(opts)
	if err != nil {
		return nil, err
	}
	return db_client.EstablishConnection(ctx, p)
	//const (
	//	maxOpenConnections = 1
	//	connMaxIdleTime    = 1 * time.Minute
	//	connMaxLifetime    = 10 * time.Minute
	//)
	//,
	//maxOpenConnections,
	//	connMaxLifetime,
	//	connMaxIdleTime

	log.Println("[TRACE] Connection string: ", psqlInfo)

	// connect to the database using the postgres driver
	utils.LogTime("db.createLocalDbClient connection open start")
	dbPool, err := pgxpool.Connect(context.Background(), psqlInfo)

	utils.LogTime("db.createLocalDbClient connection open end")

	if err != nil {
		return nil, err
	}

	if err := db_common.WaitForConnection(ctx, dbPool); err != nil {
		return nil, err
	}
	return dbPool, nil
}
