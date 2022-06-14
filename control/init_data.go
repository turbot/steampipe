package control

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v3/instrument"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controldisplay"
	"github.com/turbot/steampipe/db/db_client"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/modinstaller"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/workspace"
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

	if err := controldisplay.EnsureTemplates(); err != nil {
		initData.Result.Error = err
		return initData
	}

	// initialise telemetry
	shutdownTelemetry, err := instrument.Init(constants.AppName)
	if err != nil {
		initData.Result.AddWarnings(err.Error())
	} else {
		initData.ShutdownTelemetry = shutdownTelemetry
	}

	if viper.GetBool(constants.ArgModInstall) {
		opts := &modinstaller.InstallOpts{WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir)}
		_, err := modinstaller.InstallWorkspaceDependencies(opts)
		if err != nil {
			initData.Result.Error = err
			return initData
		}
	}

	if viper.GetString(constants.ArgOutput) == constants.CheckOutputFormatNone {
		// set progress to false
		viper.Set(constants.ArgProgress, false)
	}

	cloudMetadata, err := cmdconfig.GetCloudMetadata()
	if err != nil {
		initData.Result.Error = err
		return initData
	}

	// set cloud metadata (may be nil)
	initData.Workspace.CloudMetadata = cloudMetadata
	// set color schema
	err = initialiseColorScheme()
	if err != nil {
		initData.Result.Error = err
		return initData
	}

	// check if the required plugins are installed
	err = initData.Workspace.CheckRequiredPluginsInstalled()
	if err != nil {
		initData.Result.Error = err
		return initData
	}

	if len(initData.Workspace.GetResourceMaps().Controls) == 0 {
		initData.Result.AddWarnings("no controls found in current workspace")
	}

	statushooks.SetStatus(ctx, "Connecting to service...")
	// get a client
	var client db_common.Client
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		client, err = db_client.NewDbClient(ctx, connectionString)
	} else {
		// when starting the database, installers may trigger their own spinners
		client, err = db_local.GetLocalClient(ctx, constants.InvokerCheck)
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
	if i.Client != nil {
		i.Client.Close(ctx)
	}

	if i.ShutdownTelemetry != nil {
		i.ShutdownTelemetry()
	}
}

func initialiseColorScheme() error {
	theme := viper.GetString(constants.ArgTheme)
	if !viper.GetBool(constants.ConfigKeyIsTerminalTTY) {
		// enforce plain output for non-terminals
		theme = "plain"
	}
	themeDef, ok := controldisplay.ColorSchemes[theme]
	if !ok {
		return fmt.Errorf("invalid theme '%s'", theme)
	}
	scheme, err := controldisplay.NewControlColorScheme(themeDef)
	if err != nil {
		return err
	}
	controldisplay.ControlColors = scheme
	return nil
}
