package cmdconfig

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/db_common"
	"github.com/turbot/pipe-fittings/statushooks"
	"github.com/turbot/steampipe/pkg/db/db_local"
	pb "github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/proto"
)

func EnsureService(ctx context.Context, invoker constants.Invoker) db_common.InitResult {
	var res = db_common.InitResult{}
	// add a message rendering function to the context - this is used for the fdw update message and
	// allows us to render it as a standard initialisation message
	serviceCtx := statushooks.AddMessageRendererToContext(ctx, func(format string, a ...any) {
		res.AddMessage(fmt.Sprintf(format, a...))
	})

	// start a service if necessary
	// TODO kai check this and use a const if correct
	wd := viper.GetString(constants.ArgWorkspaceDatabase)
	if wd != "local" {
		return res
	}
	startResult := db_local.EnsureService(serviceCtx, invoker)
	res.ErrorAndWarnings = startResult.ErrorAndWarnings
	if res.Error != nil {
		return res
	}

	// after creating the client, refresh connections
	// NOTE: we cannot do this until after creating the client to ensure we do not miss notifications
	if startResult.Status == db_local.ServiceStarted {
		// ask the plugin manager to refresh connections
		// this is executed asyncronously by the plugin manager
		// we ignore this error, since RefreshConnections is async and all errors will flow through
		// the notification system
		// we do not expect any I/O errors on this since the PluginManager is running in the same box
		_, _ = startResult.PluginManager.RefreshConnections(&pb.RefreshConnectionsRequest{})
	}

	// set the connection string
	viper.Set(constants.ArgConnectionString, startResult.ConnectionString())
	return res
}
