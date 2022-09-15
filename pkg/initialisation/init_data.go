package initialisation

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v4/telemetry"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_client"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/modinstaller"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/workspace"
)

type InitData struct {
	Workspace         *workspace.Workspace
	Client            db_common.Client
	Result            *db_common.InitResult
	ShutdownTelemetry func()
}

func NewInitData(ctx context.Context, w *workspace.Workspace, invoker constants.Invoker) *InitData {
	initData := &InitData{
		Workspace: w,
		Result:    &db_common.InitResult{},
	}

	defer func() {
		// if there is no error, return context cancellation error (if any)
		if initData.Result.Error == nil {
			initData.Result.Error = ctx.Err()
		}
	}()
	// initialise telemetry
	shutdownTelemetry, err := telemetry.Init(constants.AppName)
	if err != nil {
		initData.Result.AddWarnings(err.Error())
	} else {
		initData.ShutdownTelemetry = shutdownTelemetry
	}

	// install mod dependencies if needed
	if viper.GetBool(constants.ArgModInstall) {
		opts := &modinstaller.InstallOpts{WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir)}
		_, err := modinstaller.InstallWorkspaceDependencies(opts)
		if err != nil {
			initData.Result.Error = err
			return initData
		}
	}

	// retrieve cloud metadata
	cloudMetadata, err := cmdconfig.GetCloudMetadata()
	if err != nil {
		initData.Result.Error = err
		return initData
	}

	// set cloud metadata (may be nil)
	initData.Workspace.CloudMetadata = cloudMetadata

	// check if the required plugins are installed
	err = initData.Workspace.CheckRequiredPluginsInstalled()
	if err != nil {
		initData.Result.Error = err
		return initData
	}

	// setup the session data - prepared statements and introspection tables
	sessionDataSource := workspace.NewSessionDataSource(initData.Workspace, nil)
	// define db connection callback function
	ensureSessionData := func(ctx context.Context, conn *pgx.Conn) error {
		err, preparedStatementFailures := workspace.EnsureSessionData(ctx, sessionDataSource, conn)
		w.HandlePreparedStatementFailures(preparedStatementFailures)
		return err
	}

	// get a client
	statushooks.SetStatus(ctx, "Connecting to service...")
	client, err := GetDbClient(ctx, invoker, ensureSessionData)
	if err != nil {
		initData.Result.Error = err
		return initData
	}
	initData.Client = client
	statushooks.Done(ctx)

	// refresh connections
	refreshResult := initData.Client.RefreshConnectionAndSearchPaths(ctx)
	if refreshResult.Error != nil {
		initData.Result.Error = refreshResult.Error
		return initData
	}
	// add refresh connection warnings
	initData.Result.AddWarnings(refreshResult.Warnings...)
	// add warnings from prepared statement creation
	initData.Result.AddPreparedStatementFailures(w.GetPreparedStatementFailures())
	return initData
}

// GetDbClient either creates a DB client using the configured connection string (if present) or creates a LocalDbClient
func GetDbClient(ctx context.Context, invoker constants.Invoker, onConnectionCallback db_client.DbConnectionCallback) (client db_common.Client, err error) {
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		client, err = db_client.NewDbClient(ctx, connectionString, onConnectionCallback)
	} else {
		client, err = db_local.GetLocalClient(ctx, invoker, onConnectionCallback)
	}
	return client, err
}

func (i InitData) Cleanup(ctx context.Context) {
	if i.Client != nil {
		i.Client.Close(ctx)
	}

	if i.ShutdownTelemetry != nil {
		i.ShutdownTelemetry()
	}
	if i.Workspace != nil {
		i.Workspace.Close()
	}
}
