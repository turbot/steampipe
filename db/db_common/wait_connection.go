package db_common

import (
	"context"
	"database/sql"
	"time"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// WaitForConnection waits for the db to start accepting connections and returns true
// returns false if the dbClient does not start within a stipulated time,
func WaitForConnection(ctx context.Context, db *sql.DB) (err error) {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	pingTimer := time.NewTicker(constants.ServicePingInterval)
	timeoutAt := time.After(constants.ServiceStartTimeout)
	defer pingTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pingTimer.C:
			err = db.Ping()
			if err == nil {
				return
			}
		case <-timeoutAt:
			return
		}
	}
}
