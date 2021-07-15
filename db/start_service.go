package db

import (
	"log"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// StartImplicitService starts up the service in an implicit mode
func StartImplicitService(invoker Invoker, refreshConnections bool) error {
	utils.LogTime("db.StartImplicitService start")
	defer utils.LogTime("db.StartImplicitService end")

	log.Println("[TRACE] start implicit service")

	if _, err := StartDB(constants.DatabaseDefaultPort, ListenTypeLocal, invoker, refreshConnections); err != nil {
		return err
	}
	return nil
}
