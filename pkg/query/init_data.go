package query

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_client"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/workspace"
)

type InitData struct {
	initialisation.InitData

	cancelInitialisation context.CancelFunc
	Loaded               chan struct{}
	// map of query name to resolved query (key is the query text for command line queries)
	Queries []*modconfig.ResolvedQuery
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

	statushooks.SetStatus(ctx, "Loading workspace")

	// load workspace variables syncronously
	w, errAndWarnings := workspace.LoadWorkspaceVars(ctx)
	if errAndWarnings.GetError() != nil {
		i.Result.Error = fmt.Errorf("failed to load workspace: %s", error_helpers.HandleCancelError(errAndWarnings.GetError()).Error())
		return i
	}

	i.Result.AddWarnings(errAndWarnings.Warnings...)
	i.Workspace = w

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
	// the client is set after the cancel hits
	<-i.Loaded

	// if a client was initialised, close it
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

	// load the workspace mod (this load is asynchronous as it is within the async init function)
	errAndWarnings := i.Workspace.LoadWorkspaceMod(ctx)
	i.Result.AddWarnings(errAndWarnings.Warnings...)
	if errAndWarnings.GetError() != nil {
		i.Result.Error = fmt.Errorf("failed to load workspace mod: %s", error_helpers.HandleCancelError(errAndWarnings.GetError()).Error())
		return
	}

	// set max DB connections to 1
	viper.Set(constants.ArgMaxParallel, 1)

	statushooks.SetStatus(ctx, "Resolving arguments")

	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	resolvedQueries, err := i.Workspace.GetQueriesFromArgs(args)
	if err != nil {
		i.Result.Error = err
		return
	}
	// create a cancellable context so that we can cancel the initialisation
	ctx, cancel := context.WithCancel(ctx)
	// and store it
	i.cancelInitialisation = cancel
	i.Queries = resolvedQueries

	// and call base init
	i.InitData.Init(
		ctx,
		constants.InvokerQuery,
		db_client.WithUserPoolOverride(db_client.PoolOverrides{
			Size:        1,
			MaxLifeTime: 24 * time.Hour,
			MaxIdleTime: 24 * time.Hour,
		}),
		db_client.WithManagementPoolOverride(db_client.PoolOverrides{
			// we need two connections here, since one of them will be reserved
			// by the notification listener in the interactive prompt
			Size: 2,
		}),
	)
}
