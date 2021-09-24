package db_common

import (
	"database/sql"
	"time"

	"github.com/turbot/steampipe/utils"
)

// waits for the db to start accepting connections and returns true
// returns false if the dbClient does not start within a stipulated time,
func WaitForConnection(pool *sql.DB) bool {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	pingTimer := time.NewTicker(10 * time.Millisecond)
	timeoutAt := time.After(5 * time.Second)
	defer pingTimer.Stop()
	for {
		select {
		case <-pingTimer.C:
			pingErr := pool.Ping()
			if pingErr == nil {
				return true
			}
		case <-timeoutAt:
			return false
		}
	}
}
