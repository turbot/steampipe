package initialisation

import (
	"context"
	"fmt"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v4/telemetry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_client"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/modinstaller"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/workspace"
)

type InitData struct {
	Workspace *workspace.Workspace
	Client    db_common.Client
	Result    *db_common.InitResult

	// used for query only
	PreparedStatementSource *modconfig.ModResources

	ShutdownTelemetry func()
}

func NewInitData(ctx context.Context, w *workspace.Workspace) *InitData {
	i := &InitData{
		Workspace: w,
		Result:    &db_common.InitResult{},
	}
	i.Init(ctx)
	return i
}

func (i *InitData) Init(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			i.Result.Error = helpers.ToError(r)
		}
		// if there is no error, return context cancellation error (if any)
		if i.Result.Error == nil {
			i.Result.Error = ctx.Err()
		}
	}()

	// initialise telemetry
	shutdownTelemetry, err := telemetry.Init(constants.AppName)
	if err != nil {
		i.Result.AddWarnings(err.Error())
	} else {
		i.ShutdownTelemetry = shutdownTelemetry
	}

	// install mod dependencies if needed
	if viper.GetBool(constants.ArgModInstall) {
		opts := &modinstaller.InstallOpts{WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir)}
		_, err := modinstaller.InstallWorkspaceDependencies(opts)
		if err != nil {
			i.Result.Error = err
			return
		}
	}

	// retrieve cloud metadata
	cloudMetadata, err := getCloudMetadata()
	if err != nil {
		i.Result.Error = err
		return
	}

	// set cloud metadata (may be nil)
	i.Workspace.CloudMetadata = cloudMetadata

	// check if the required plugins are installed
	err = i.Workspace.CheckRequiredPluginsInstalled()
	if err != nil {
		i.Result.Error = err
		return
	}

	//validate steampipe version
	if err = i.Workspace.ValidateSteampipeVersion(); err != nil {
		i.Result.Error = err
		return
	}

	// get a client

	// add a message rendering function to the context - this is used for the fdw update message and
	// allows us to render it as a standard initialisation message
	getClientCtx := statushooks.AddMessageRendererToContext(ctx, func(format string, a ...any) {
		i.Result.AddMessage(fmt.Sprintf(format, a...))
	})

	client, err := i.getClient(getClientCtx, err)
	if err != nil {
		i.Result.Error = err
		return
	}
	i.Client = client

	// refresh connections
	refreshResult := i.Client.RefreshConnectionAndSearchPaths(ctx)
	if refreshResult.Error != nil {
		i.Result.Error = refreshResult.Error
		return
	}
	i.Result.AddWarnings(refreshResult.Warnings...)

	// setup the session data - prepared statements and introspection tables
	sessionDataSource := workspace.NewSessionDataSource(i.Workspace, i.PreparedStatementSource)

	// register EnsureSessionData as a callback on the client.
	// if the underlying SQL client has certain errors (for example context expiry) it will reset the session
	// so our client object calls this callback to restore the session data
	i.Client.SetEnsureSessionDataFunc(func(localCtx context.Context, conn *db_common.DatabaseSession) (error, []string) {
		return workspace.EnsureSessionData(localCtx, sessionDataSource, conn)
	})

	// force creation of session data - se we see any prepared statement errors at once
	sessionResult := i.Client.AcquireSession(ctx)
	i.Result.AddWarnings(sessionResult.Warnings...)
	if sessionResult.Error != nil {
		i.Result.Error = fmt.Errorf("error acquiring database connection, %s", sessionResult.Error.Error())
	} else {
		sessionResult.Session.Close(utils.IsContextCancelled(ctx))
	}

	return
}

func (i *InitData) getClient(ctx context.Context, err error) (db_common.Client, error) {
	statushooks.SetStatus(ctx, "Connecting to service...")
	defer statushooks.Done(ctx)
	var client db_common.Client
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		client, err = db_client.NewDbClient(ctx, connectionString)
	} else {
		// when starting the database, installers may trigger their own spinners
		client, err = db_local.GetLocalClient(ctx, constants.InvokerDashboard)
	}
	return client, err
}

func (i *InitData) Cleanup(ctx context.Context) {
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
