package query

import (
	"context"
	"fmt"
	localcmdconfig "github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/db/steampipe_db_client"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/db_client"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/export"
	"github.com/turbot/pipe-fittings/initialisation"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/statushooks"
	"github.com/turbot/pipe-fittings/workspace"
)

type InitData struct {
	initialisation.InitData

	cancelInitialisation context.CancelFunc
	Loaded               chan struct{}
	// map of query name to resolved query (key is the query text for command line queries)
	Queries map[string]*modconfig.ResolvedQuery
	Client  *steampipe_db_client.SteampipeDbClient
}

// NewInitData returns a new InitData object
// It also starts an asynchronous population of the object
// InitData.Done closes after asynchronous initialization completes
func NewInitData(ctx context.Context, args []string) *InitData {
	i := &InitData{
		InitData: *initialisation.NewInitData(),
		Loaded:   make(chan struct{}),
	}
	// for interactive mode - do the home directory modfile check before init
	if viper.GetBool(constants.ConfigKeyInteractive) {
		path := viper.GetString(constants.ArgModLocation)
		modFilePath, _ := workspace.FindModFilePath(path)

		// if the user cancels - no need to continue init
		if err := workspace.HomeDirectoryModfileCheck(ctx, filepath.Dir(modFilePath)); err != nil {
			i.Result.Error = err
			close(i.Loaded)
			return i
		}
		// home dir modfile already done - set the viper config
		viper.Set(constants.ConfigKeyBypassHomeDirModfileWarning, true)
	}
	go i.init(ctx, args)

	return i
}

func queryExporters() []export.Exporter {
	return []export.Exporter{&export.SnapshotExporter{}}
}

func (i *InitData) Cancel() {
	// cancel any ongoing operation
	if i.cancelInitialisation != nil {
		i.cancelInitialisation()
	}
	i.cancelInitialisation = nil
}

// Cleanup overrides the initialisation.InitData.Cleanup to provide syncronisation with the loaded channel
func (i *InitData) Cleanup(ctx context.Context) {
	// cancel any ongoing operation
	i.Cancel()

	// ensure that the initialisation was completed
	// and that we are not in a race condition where
	// the Client is set after the cancel hits
	<-i.Loaded

	// if a Client was initialised, close it
	if i.Client != nil {
		i.Client.Close(ctx)
	}
	if i.ShutdownTelemetry != nil {
		i.ShutdownTelemetry()
	}
}

func (i *InitData) init(ctx context.Context, args []string) {
	defer func() {
		close(i.Loaded)
		// clear the cancelInitialisation function
		i.cancelInitialisation = nil
	}()

	// validate export args
	if len(viper.GetStringSlice(constants.ArgExport)) > 0 {
		i.RegisterExporters(queryExporters()...)

		// validate required export formats
		if err := i.ExportManager.ValidateExportFormat(viper.GetStringSlice(constants.ArgExport)); err != nil {
			i.Result.Error = err
			return
		}
	}

	statushooks.SetStatus(ctx, "Loading workspace")
	w, errAndWarnings := workspace.LoadWorkspacePromptingForVariables(ctx)
	if errAndWarnings.GetError() != nil {
		i.Result.Error = fmt.Errorf("failed to load workspace: %s", error_helpers.HandleCancelError(errAndWarnings.GetError()).Error())
		return
	}
	i.Result.AddWarnings(errAndWarnings.Warnings...)
	i.Workspace = w

	// set max DB connections to 1
	viper.Set(constants.ArgMaxParallel, 1)

	statushooks.SetStatus(ctx, "Resolving arguments")

	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	resolvedQueries, err := w.GetQueriesFromArgs(args)
	if err != nil {
		i.Result.Error = err
		return
	}
	// create a cancellable context so that we can cancel the initialisation
	ctx, cancel := context.WithCancel(ctx)
	// and store it
	i.cancelInitialisation = cancel
	i.Queries = resolvedQueries

	ew := localcmdconfig.EnsureService(ctx, constants.InvokerQuery)
	if ew.GetError() != nil {
		i.Result.Error = ew.Error
		return
	}
	i.Result.AddWarnings(ew.Warnings...)

	// TODO KAI DO POOL OVERRIDES STILL WORK?
	// and call base init
	i.InitData.Init(ctx,
		db_client.WithUserPoolOverride(db_client.PoolOverrides{
			Size:        1,
			MaxLifeTime: 24 * time.Hour,
			MaxIdleTime: 24 * time.Hour,
		}),
		db_client.WithManagementPoolOverride(db_client.PoolOverrides{
			// we need two connections here, since one of them will be reserved
			// by the notification listener in the interactive prompt
			Size: 2,
		}))

	// TODO KAI OnConnectionCallback
	// now wrap the Client
	if i.Result.Error == nil {
		i.Client, i.Result.Error = steampipe_db_client.NewSteampipeDbClient(ctx, i.InitData.Client, nil)
	}
}
