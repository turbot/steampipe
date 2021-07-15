package db

import (
	"log"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// StartImplicitService starts up the service in an implicit mode
func StartImplicitService(invoker Invoker) error {
	utils.LogTime("db.StartImplicitService start")
	defer utils.LogTime("db.StartImplicitService end")

	log.Println("[TRACE] start implicit service")

	if _, err := StartDB(constants.DatabaseDefaultPort, ListenTypeLocal, invoker, false); err != nil {
		return err
	}
	return nil
}
