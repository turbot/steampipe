package initialisation

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/telemetry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_client"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/modinstaller"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/workspace"
)

type InitData struct {
	Workspace *workspace.Workspace
	Client    db_common.Client
	Result    *db_common.InitResult

	ShutdownTelemetry func()
	ExportManager     *export.Manager
	ConnectionMap     steampipeconfig.ConnectionDataMap
}

func NewErrorInitData(err error) *InitData {
	return &InitData{
		Result: &db_common.InitResult{Error: err},
	}
}

func NewInitData() *InitData {
	i := &InitData{
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

func (i *InitData) Init(ctx context.Context, invoker constants.Invoker) {
	defer func() {
		if r := recover(); r != nil {
			i.Result.Error = helpers.ToError(r)
		}
		// if there is no error, return context cancellation error (if any)
		if i.Result.Error == nil {
			i.Result.Error = ctx.Err()
		}
	}()

	statushooks.SetStatus(ctx, "Initializing")

	// initialise telemetry
	shutdownTelemetry, err := telemetry.Init(constants.AppName)
	if err != nil {
		i.Result.AddWarnings(err.Error())
	} else {
		i.ShutdownTelemetry = shutdownTelemetry
	}

	// install mod dependencies if needed
	if viper.GetBool(constants.ArgModInstall) {
		statushooks.SetStatus(ctx, "Installing workspace dependencies")
		opts := &modinstaller.InstallOpts{WorkspacePath: viper.GetString(constants.ArgModLocation)}
		if _, err := modinstaller.InstallWorkspaceDependencies(opts); err != nil {
			i.Result.Error = err
		}
		return
	}

	// retrieve cloud metadata
	cloudMetadata, err := getCloudMetadata(ctx)
	if err != nil {
		i.Result.Error = err
		return
	}

	// set cloud metadata (may be nil)
	i.Workspace.CloudMetadata = cloudMetadata

	statushooks.SetStatus(ctx, "Checking for required plugins")
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

	// if introspection tables are enabled, setup the session data callback
	var ensureSessionData db_client.DbConnectionCallback
	if viper.GetString(constants.ArgIntrospection) != constants.IntrospectionNone {
		ensureSessionData = func(ctx context.Context, conn *pgx.Conn) error {
			return workspace.EnsureSessionData(ctx, i.Workspace.GetResourceMaps(), conn)
		}
	}

	// get a client
	// add a message rendering function to the context - this is used for the fdw update message and
	// allows us to render it as a standard initialisation message
	getClientCtx := statushooks.AddMessageRendererToContext(ctx, func(format string, a ...any) {
		i.Result.AddMessage(fmt.Sprintf(format, a...))
	})

	statushooks.SetStatus(ctx, "Connecting to steampipe")
	client, errorsAndWarnings := GetDbClient(getClientCtx, invoker, ensureSessionData)
	if errorsAndWarnings.Error != nil {
		i.Result.Error = errorsAndWarnings.Error
		return
	}
	i.Result.AddWarnings(errorsAndWarnings.Warnings...)
	i.Client = client
	// if the cache is set on the workspace profile
	// and is set to false, then override the default cache setting of the connection
	if viper.IsSet(constants.ArgClientCacheEnabled) && !viper.GetBool(constants.ArgClientCacheEnabled) {
		if e := i.Client.CacheOff(ctx); e != nil {
			i.Result.Error = e
			return
		}
	}
	if viper.IsSet(constants.ArgCacheTtl) {
		ttl := time.Duration(viper.GetInt(constants.ArgCacheTtl)) * time.Second
		if e := i.Client.SetCacheTtl(ctx, ttl); e != nil {
			i.Result.Error = e
			return
		}
	}

	// load the connection state and cache it!
	connectionMap, _, err := steampipeconfig.GetConnectionState(client.ForeignSchemaNames())
	if err != nil {
		i.Result.Error = err
		return
	}

	i.ConnectionMap = connectionMap

}

// GetDbClient either creates a DB client using the configured connection string (if present) or creates a LocalDbClient
func GetDbClient(ctx context.Context, invoker constants.Invoker, onConnectionCallback db_client.DbConnectionCallback) (db_common.Client, *modconfig.ErrorAndWarnings) {
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		statushooks.SetStatus(ctx, "Connecting to remote Steampipe database")
		client, err := db_client.NewDbClient(ctx, connectionString, onConnectionCallback)
		return client, modconfig.NewErrorsAndWarning(err)
	}

	statushooks.SetStatus(ctx, "Starting local Steampipe database")
	return db_local.GetLocalClient(ctx, invoker, onConnectionCallback)
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
