package db

import (
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/local_db"
)

func GetLocalClient(invoker constants.Invoker) (db_common.Client, error) {
	// start db if necessary
	err := local_db.EnsureDbAndStartService(invoker)
	if err != nil {
		return nil, err
	}

	client, err := local_db.NewLocalClient(invoker)
	if err != nil {
		local_db.ShutdownService(invoker)
	}
	// NOTE:  client shutdown will shutdown service (if invoker matches)
	return client, nil
}
