package initialisation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/telemetry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_client"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/modinstaller"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/workspace"
)

type InitData struct {
	// the current state that init is in
	Status string
	// if non-nil, this is called everytime the status changes
	OnStatusChanged func(string)

	Workspace *workspace.Workspace
	Client    db_common.Client
	Result    *db_common.InitResult

	// used for query only
	PreparedStatementSource *modconfig.ResourceMaps

	ShutdownTelemetry func()
	ExportManager     *export.Manager
}

func (i *InitData) SetStatus(newStatus string) {
	i.Status = newStatus
	if i.OnStatusChanged != nil {
		i.OnStatusChanged(newStatus)
	}
}

func NewErrorInitData(err error) *InitData {
	return &InitData{
		Result: &db_common.InitResult{Error: err},
	}
}

func NewInitData(w *workspace.Workspace) *InitData {
	i := &InitData{
		Workspace:     w,
		Result:        &db_common.InitResult{},
		ExportManager: export.NewManager(),
	}

	return i
}

func (i *InitData) RegisterExporters(exporters ...export.Exporter) *InitData {
	for _, e := range exporters {
		i.ExportManager.Register(e)
	}

	return i
}

func (i *InitData) Init(parentCtx context.Context, invoker constants.Invoker) (res *InitData) {
	// create a context with the init hook in - which can be sent down to lower level operations
	hook := NewInitStatusHook(i)
	ctx := statushooks.AddStatusHooksToContext(parentCtx, hook)

	defer func() {
		if r := recover(); r != nil {
			i.Result.Error = helpers.ToError(r)
		}
		// if there is no error, return context cancellation error (if any)
		if i.Result.Error == nil {
			i.Result.Error = ctx.Err()
		}
	}()
	// return ourselves
	res = i

	i.SetStatus("Initializing")

	// initialise telemetry
	shutdownTelemetry, err := telemetry.Init(constants.AppName)
	if err != nil {
		i.Result.AddWarnings(err.Error())
	} else {
		i.ShutdownTelemetry = shutdownTelemetry
	}

	// install mod dependencies if needed
	if viper.GetBool(constants.ArgModInstall) {
		i.SetStatus("Installing workspace dependencies")
		opts := &modinstaller.InstallOpts{WorkspacePath: viper.GetString(constants.ArgModLocation)}
		_, err := modinstaller.InstallWorkspaceDependencies(ctx, opts)
		if err != nil {
			i.Result.Error = err
			return
		}
	}

	// retrieve cloud metadata
	cloudMetadata, err := getCloudMetadata(ctx)
	if err != nil {
		i.Result.Error = err
		return
	}

	// set cloud metadata (may be nil)
	i.Workspace.CloudMetadata = cloudMetadata

	i.SetStatus("Checking for required plugins")
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
	sessionDataSource := workspace.NewSessionDataSource(i.Workspace, i.PreparedStatementSource)
	// define db connection callback function
	ensureSessionData := func(ctx context.Context, conn *pgx.Conn) error {
		return workspace.EnsureSessionData(ctx, sessionDataSource, conn)
	}

	// get a client
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

	// force creation of session data - se we see any prepared statement errors at once
	sessionResult := i.Client.AcquireSession(ctx)
	i.Result.AddWarnings(sessionResult.Warnings...)
	if sessionResult.Error != nil {
		i.Result.Error = fmt.Errorf("error acquiring database connection, %s", sessionResult.Error.Error())
	} else {
		sessionResult.Session.Close(error_helpers.IsContextCanceled(ctx))
	}
	// add refresh connection warnings
	i.Result.AddWarnings(refreshResult.Warnings...)

	return
}

// GetDbClient either creates a DB client using the configured connection string (if present) or creates a LocalDbClient
func GetDbClient(ctx context.Context, invoker constants.Invoker, onConnectionCallback db_client.DbConnectionCallback) (client db_common.Client, err error) {
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		statushooks.SetStatus(ctx, "Connecting to Remote Steampipe")
		client, err = db_client.NewDbClient(ctx, connectionString, onConnectionCallback)
	} else {
		statushooks.SetStatus(ctx, "Starting local Steampipe")
		client, err = db_local.GetLocalClient(ctx, invoker, onConnectionCallback)
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
