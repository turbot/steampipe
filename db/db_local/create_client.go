package db_local

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/utils"
)

func getLocalSteampipeConnectionString() (string, error) {
	utils.LogTime("db.createDbClient start")
	defer utils.LogTime("db.createDbClient end")
	log.Println("[TRACE] createDbClient")

	info, err := GetStatus()
	if err != nil {
		return "", err
	}
	if info == nil {
		return "", fmt.Errorf("steampipe service is not running")
	}

	// Connect to the database using the first listen address, which is usually localhost
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", info.Listen[0], info.Port, constants.DatabaseUser, info.Database, SslMode())

	return psqlInfo, nil
}

// createRootDbClient connects as a superuser to the
// installed database, if available, otherwise to the default
// "postgres" database
func createRootDbClient() (*sql.DB, error) {
	utils.LogTime("db.createSteampipeRootDbClient start")
	defer utils.LogTime("db™.createSteampipeRootDbClient end")

	return createLocalDbClient("", constants.DatabaseSuperUser)
}

// createLocalDbClient connects and returns a connection to the given database using
// the provided username
// if the database is not provided, then it fallsback to the default database in the service
// that was created during installation.
func createLocalDbClient(dbname string, username string) (*sql.DB, error) {
	utils.LogTime("db.createDbClient start")
	defer utils.LogTime("db.createDbClient end")

	info, err := GetStatus()

	if err != nil {
		return nil, err
	}

	if info == nil {
		return nil, fmt.Errorf("steampipe service is not running")
	}

	if len(dbname) == 0 {
		dbname = info.Database
	}
	// if we still don't have it, fallback to default "postgres"
	if len(dbname) == 0 {
		dbname = "postgres"
	}

	// Connect to the database using the first listen address, which is usually localhost
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", info.Listen[0], info.Port, username, dbname, SslMode())

	log.Println("[TRACE] status: ", info)
	log.Println("[TRACE] Connection string: ", psqlInfo)

	// connect to the database using the postgres driver
	utils.LogTime("db.createDbClient connection open start")
	db, err := sql.Open("postgres", psqlInfo)
	db.SetMaxOpenConns(1)
	utils.LogTime("db.createDbClient connection open end")

	if err != nil {
		return nil, err
	}

	if db_common.WaitForConnection(db) {
		return db, nil
	}

	return nil, fmt.Errorf("could not establish connection with database")
}
