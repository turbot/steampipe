package local_db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

func createSteampipeDbClient() (*sql.DB, error) {
	utils.LogTime("db.createSteampipeDbClient start")
	defer utils.LogTime("db.createSteampipeDbClient end")

	return createDbClient(constants.DatabaseName, constants.DatabaseUser)
}

func createRootDbClient() (*sql.DB, error) {
	utils.LogTime("db.createSteampipeRootDbClient start")
	defer utils.LogTime("db.createSteampipeRootDbClient end")

	return createDbClient(constants.DatabaseName, constants.DatabaseSuperUser)
}

func createDbClient(dbname string, username string) (*sql.DB, error) {
	utils.LogTime("db.createDbClient start")
	utils.LogTime(fmt.Sprintf("to %s with %s", dbname, username))
	defer utils.LogTime("db.createDbClient end")

	log.Println("[TRACE] createDbClient")
	info, err := GetStatus()

	if err != nil {
		return nil, err
	}

	if info == nil {
		return nil, fmt.Errorf("steampipe service is not running")
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

	if waitForConnection(db) {
		return db, nil
	}

	return nil, fmt.Errorf("could not establish connection with database")
}

func createDbClientWithConnectionString(connectionString string) (*sql.DB, error) {

	utils.LogTime("db.createDbClientWithConnectionString start")
	defer utils.LogTime("db.createDbClient end")

	log.Println("[TRACE] createDbClient")

	// TODO add in username and password
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	// NEEDED?
	if waitForConnection(db) {
		return db, nil
	}

	return nil, fmt.Errorf("could not establish connection with database")
}
