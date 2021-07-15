package db

import (
	"log"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// StartImplicitService :: invokes `steampipe service start --database-listen local --refresh=false --invoker query`
func StartImplicitService(invoker Invoker, refreshConnections bool) {
	utils.LogTime("db.StartImplicitService start")
	defer utils.LogTime("db.StartImplicitService end")

	log.Println("[TRACE] start implicit service")

	// start db but DO NOT refresh connections - this will be done explicitly later
	StartDB(constants.DatabaseDefaultPort, ListenTypeLocal, invoker, refreshConnections)
}
