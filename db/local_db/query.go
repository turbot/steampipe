package local_db

import (
	"errors"
	"log"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/utils"
)

// EnsureDbAndStartService :: ensure db is installed and start service if necessary
func EnsureDbAndStartService(invoker constants.Invoker) error {
	utils.LogTime("db.EnsureDbAndStartService start")
	defer utils.LogTime("db.EnsureDbAndStartService end")

	log.Println("[TRACE] db.EnsureDbAndStartService start")

	if err := EnsureDBInstalled(); err != nil {
		return err
	}

	status, err := GetStatus()
	if err != nil {
		return errors.New("could not retrieve service status")
	}

	if status == nil {
		// the db service is not started - start it
		utils.LogTime("StartImplicitService start")
		log.Println("[TRACE] start implicit service")

		if _, err := StartDB(constants.DatabaseDefaultPort, ListenTypeLocal, invoker); err != nil {
			return err
		}
		utils.LogTime("StartImplicitService end")

		return nil

	}
	return nil
}
