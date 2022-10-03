package initialisation

import (
	"context"
	"fmt"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"

	"fmt"
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
	Workspace            *workspace.Workspace
	Client               db_common.Client
	Result               *db_common.InitResult
	cancelInitialisation context.CancelFunc
	// used for query only
	PreparedStatementSource *modconfig.ModResources

	ShutdownTelemetry func()
}

func NewInitData(ctx context.Context, w *workspace.Workspace, invoker constants.Invoker) *InitData {
	i := &InitData{
		Workspace: w,
		Result:    &db_common.InitResult{},
	}
	i.Init(ctx, invoker)
	return i
}

func (i *InitData) Init(ctx context.Context, invoker constants.Invoker) {
	defer func() {
		if r := recover(); r != nil {
			i.Result.Error = helpers.ToError(r)
		}
		// if there is no error, return context cancellation error (if any)
		if i.Result.Error == nil {
			i.Result.Error = ctx.Err()
		}
		// clear the cancelInitialisation function
		i.cancelInitialisation = nil
	}()

	// create a cancellable context so that we can cancel the initialisation
	ctx, cancel := context.WithCancel(ctx)
	// and store it
	i.cancelInitialisation = cancel

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
	cloudMetadata, err := cmdconfig.GetCloudMetadata()
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

	// setup the session data - prepared statements and introspection tables
	sessionDataSource := workspace.NewSessionDataSource(i.Workspace, nil)
	// define db connection callback function
	ensureSessionData := func(ctx context.Context, conn *pgx.Conn) error {
		err, preparedStatementFailures := workspace.EnsureSessionData(ctx, sessionDataSource, conn)
		w.HandlePreparedStatementFailures(preparedStatementFailures)
		return err
	}

	// get a client
	statushooks.SetStatus(ctx, "Connecting to service...")
	// add a message rendering function to the context - this is used for the fdw update message and
	// allows us to render it as a standard initialisation message
	getClientCtx := statushooks.AddMessageRendererToContext(ctx, func(format string, a ...any) {
		i.Result.AddMessage(fmt.Sprintf(format, a...))
	})

	client, err := GetDbClient(getClientCtx, invoker, ensureSessionData)
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
	// add refresh connection warnings
	i.Result.AddWarnings(refreshResult.Warnings...)
	// add warnings from prepared statement creation
	i.Result.AddPreparedStatementFailures(w.GetPreparedStatementFailures())

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
func (i *InitData) Cancel() {
	// cancel any ongoing operation
	if i.cancelInitialisation != nil {
		i.cancelInitialisation()
	}
	i.cancelInitialisation = nil
}


// GetDbClient either creates a DB client using the configured connection string (if present) or creates a LocalDbClient
func GetDbClient(ctx context.Context, invoker constants.Invoker, onConnectionCallback db_client.DbConnectionCallback) (client db_common.Client, err error) {
	statushooks.SetStatus(ctx, "Connecting to service...")
	defer statushooks.Done(ctx)

	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		client, err = db_client.NewDbClient(ctx, connectionString, onConnectionCallback)
	} else {
		client, err = db_local.GetLocalClient(ctx, invoker, onConnectionCallback)
	}
	return client, err
}

func (i InitData) Cleanup(ctx context.Context) {
	// cancel any ongoing operation
	i.Cancel()

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
