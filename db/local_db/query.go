package local_db

import (
	"errors"
	"log"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/utils"
)

// EnsureDbAndStartService :: ensure db is installed and start service if necessary
func EnsureDbAndStartService(invoker constants.Invoker, refreshConnections bool) error {
	utils.LogTime("db.EnsureDbAndStartService start")
	defer utils.LogTime("db.EnsureDbAndStartService end")

	log.Println("[TRACE] db.EnsureDbAndStartService start")

	EnsureDBInstalled()
	status, err := GetStatus()
	if err != nil {
		return errors.New("could not retrieve service status")
	}

	if status == nil {
		// the db service is not started - start it
		return StartImplicitService(invoker, refreshConnections)
	}
	return nil
}
