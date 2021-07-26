package db

import (
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/local_db"
)

func GetClient(invoker constants.Invoker) (db_common.Client, error) {
	// start db if necessary, refreshing connections
	err := local_db.EnsureDbAndStartService(invoker)
	if err != nil {
		// TODO ensure source errors are complete and do not need prefixing
		//if !utils.IsCancelledError(err) {
		//	err = utils.PrefixError(err, "failed to start service")
		//}

		return nil, err
	}

	client, err := local_db.NewLocalClient(invoker)
	if err != nil {
		local_db.ShutdownService(invoker)
	}
	// NOTE:  client shutdown will shutdown service (if invoker matches)
	return client, nil
}
