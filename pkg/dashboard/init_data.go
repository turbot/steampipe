package dashboard

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v3/telemetry"

	"github.com/spf13/viper"
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

func NewInitData(ctx context.Context, w *workspace.Workspace) *InitData {
	initData := &InitData{
		Workspace: w,
		Result:    &db_common.InitResult{},
	}

	// initialise telemetry
	shutdownTelemetry, err := telemetry.Init(constants.AppName)
	if err != nil {
		initData.Result.AddWarnings(err.Error())
	} else {
		initData.ShutdownTelemetry = shutdownTelemetry
	}

	if !w.ModfileExists() {
		initData.Result.Error = workspace.ErrorNoModDefinition
		return initData
	}

	if viper.GetBool(constants.ArgModInstall) {
		opts := &modinstaller.InstallOpts{WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir)}
		_, err := modinstaller.InstallWorkspaceDependencies(opts)
		if err != nil {
			initData.Result.Error = err
			return initData
		}
	}
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

	statushooks.SetStatus(ctx, "Connecting to service...")
	// get a client
	var client db_common.Client
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		client, err = db_client.NewDbClient(ctx, connectionString)
	} else {
		// when starting the database, installers may trigger their own spinners
		client, err = db_local.GetLocalClient(ctx, constants.InvokerDashboard)
	}

	if err != nil {
		initData.Result.Error = err
		return initData
	}
	initData.Client = client

	refreshResult := initData.Client.RefreshConnectionAndSearchPaths(ctx)
	if refreshResult.Error != nil {
		initData.Result.Error = refreshResult.Error
		return initData
	}
	initData.Result.AddWarnings(refreshResult.Warnings...)

	// setup the session data - prepared statements and introspection tables
	sessionDataSource := workspace.NewSessionDataSource(initData.Workspace, nil)

	// register EnsureSessionData as a callback on the client.
	// if the underlying SQL client has certain errors (for example context expiry) it will reset the session
	// so our client object calls this callback to restore the session data
	initData.Client.SetEnsureSessionDataFunc(func(localCtx context.Context, conn *db_common.DatabaseSession) (error, []string) {
		return workspace.EnsureSessionData(localCtx, sessionDataSource, conn)
	})

	return initData

}

func (i InitData) Cleanup(ctx context.Context) {
	// if a client was initialised, close it
	if i.Client != nil {
		i.Client.Close(ctx)
	}

	if i.ShutdownTelemetry != nil {
		i.ShutdownTelemetry()
	}
}
