package initialisation

import (
	"context"
	"github.com/jackc/pgx/v4"
	"log"

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

func NewInitData(ctx context.Context, w *workspace.Workspace) *InitData {
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
		err, warnings := workspace.EnsureSessionData(ctx, sessionDataSource, conn)
		// TODO KAI how do we display wanrings
		log.Println("[WARN]", warnings)
		return err
	}

	// TODO KAI init session func
	// get a client
	statushooks.SetStatus(ctx, "Connecting to service...")
	client, err := GetDbClient(ctx, ensureSessionData)
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
	initData.Result.AddWarnings(refreshResult.Warnings...)

	// register EnsureSessionData as a callback on the client.
	// if the underlying SQL client has certain errors (for example context expiry) it will reset the session
	// so our client object calls this callback to restore the session data
	//initData.Client.SetEnsureSessionDataFunc(func(localCtx context.Context, conn *db_common.DatabaseSession) (error, []string) {
	//	return workspace.EnsureSessionData(localCtx, sessionDataSource, conn)
	//})

	return initData
}

func GetDbClient(ctx context.Context, onConnectionCallback db_client.DbConnectionCallback) (client db_common.Client, err error) {
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		client, err = db_client.NewDbClient(ctx, connectionString, onConnectionCallback)
	} else {
		// TODO KAI SORT OUT INVOKER
		// when starting the database, installers may trigger their own spinners
		client, err = db_local.GetLocalClient(ctx, constants.InvokerDashboard, onConnectionCallback)
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
