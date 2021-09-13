package local_db

import (
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
		return err
	}

	if status == nil {
		// the db service is not started - start it
		utils.LogTime("StartImplicitService start")
		log.Println("[TRACE] start implicit service")

		if _, err := StartDB(constants.DatabaseDefaultPort, ListenTypeLocal, invoker); err != nil {
			return err
		}
		utils.LogTime("StartImplicitService end")
	} else {
		// so db is already running - ensure it contains command schema
		// this is to handle the upgrade edge case where a user has a service running of an earlier version of steampipe
		// and upgrades to this version - we need to ensure we create the command schema
		return ensureCommandSchema()
	}
	return nil
}
