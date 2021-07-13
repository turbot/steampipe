package db

import (
	"log"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// StartImplicitService :: invokes `steampipe service start --database-listen local --refresh=false --invoker query`
func StartImplicitService(invoker Invoker) {
	utils.LogTime("db.StartImplicitService start")
	defer utils.LogTime("db.StartImplicitService end")

	log.Println("[TRACE] start implicit service")

	StartDB(constants.DatabaseDefaultPort, ListenTypeLocal, invoker, false)
}
