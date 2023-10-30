package localcmdconfig

import (
	"context"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/steampipe/pkg/db/db_local"
)

func EnsureService(ctx context.Context, invoker constants.Invoker) error_helpers.ErrorAndWarnings {
	// start a service if necessary
	wd := viper.GetString(constants.ArgWorkspaceDatabase)
	if wd != "local" {
		return error_helpers.ErrorAndWarnings{}
	}
	startResult := db_local.EnsureService(ctx, invoker)
	if startResult.Error != nil {
		return startResult.ErrorAndWarnings
	}
	// set the connection string
	viper.Set(constants.ArgConnectionString, startResult.ConnectionString())
	return startResult.ErrorAndWarnings
}
