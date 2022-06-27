package db_local

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
)

func getLocalSteampipeConnectionString() (string, error) {
	utils.LogTime("db.createDbClient start")
	defer utils.LogTime("db.createDbClient end")

	info, err := GetState()
	if err != nil {
		return "", err
	}
	if info == nil {
		return "", fmt.Errorf("steampipe service is not running")
	}

	// Connect to the database using the first listen address, which is usually localhost
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", info.Listen[0], info.Port, constants.DatabaseUser, info.Database, sslMode())

	return psqlInfo, nil
}

type CreateDbOptions struct {
	DatabaseName, Username string
}

// createLocalDbClient connects and returns a connection to the given database using
// the provided username
// if the database is not provided (empty), it connects to the default database in the service
// that was created during installation.
func createLocalDbClient(ctx context.Context, opts *CreateDbOptions) (*sql.DB, error) {
	utils.LogTime("db.createLocalDbClient start")
	defer utils.LogTime("db.createLocalDbClient end")

	// load the db status
	info, err := GetState()
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("steampipe service is not running")
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
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", info.Listen[0], info.Port, opts.Username, opts.DatabaseName, sslMode())

	log.Println("[TRACE] status: ", info)
	log.Println("[TRACE] Connection string: ", psqlInfo)

	// connect to the database using the postgres driver
	utils.LogTime("db.createLocalDbClient connection open start")
	db, err := sql.Open("pgx", psqlInfo)
	db.SetMaxOpenConns(1)
	utils.LogTime("db.createLocalDbClient connection open end")

	if err != nil {
		return nil, err
	}

	if err := db_common.WaitForConnection(ctx, db); err != nil {
		return nil, err
	}
	return db, nil
}
