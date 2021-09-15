package db_local

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

func createSteampipeDbClient() (*sql.DB, error) {
	utils.LogTime("db.createSteampipeDbClient start")
	defer utils.LogTime("db.createSteampipeDbClient end")

	return createLocalDbClient(constants.DatabaseName, constants.DatabaseUser)
}

func createRootDbClient() (*sql.DB, error) {
	utils.LogTime("db.createSteampipeRootDbClient start")
	defer utils.LogTime("db™.createSteampipeRootDbClient end")

	return createLocalDbClient(constants.DatabaseName, constants.DatabaseSuperUser)
}

func createLocalDbClient(dbname string, username string) (*sql.DB, error) {
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

// waits for the db to start accepting connections and returns true
// returns false if the dbClient does not start within a stipulated time,
func waitForConnection(conn *sql.DB) bool {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	pingTimer := time.NewTicker(10 * time.Millisecond)
	timeoutAt := time.After(5 * time.Second)
	defer pingTimer.Stop()
	for {
		select {
		case <-pingTimer.C:
			pingErr := conn.Ping()
			if pingErr == nil {
				return true
			}
		case <-timeoutAt:
			return false
		}
	}
}
