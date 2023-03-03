package query

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/workspace"
)

type InitData struct {
	initialisation.InitData

	// the current state that init is in
	Status string
	// if non-nil, this is called everytime the status changes
	OnStatusChanged func(string)

	cancelInitialisation context.CancelFunc
	Loaded               chan struct{}
	// map of query name to resolved query (key is the query text for command line queries)
	Queries map[string]*modconfig.ResolvedQuery
}

// NewInitData returns a new InitData object
// It also starts an asynchronous population of the object
// InitData.Done closes after asynchronous initialization completes
func NewInitData(ctx context.Context, args []string) *InitData {
	i := &InitData{
		InitData: *initialisation.NewInitData(),
		Loaded:   make(chan struct{}),
	}
	go i.init(ctx, args)

	return i
}

func (i *InitData) SetStatus(newStatus string) {
	i.Status = newStatus
	if i.OnStatusChanged != nil {
		i.OnStatusChanged(newStatus)
	}
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

func (i *InitData) init(parentCtx context.Context, args []string) {
	defer func() {
		close(i.Loaded)
		// clear the cancelInitialisation function
		i.cancelInitialisation = nil
	}()

	// create a context with the init hook in - which can be sent down to lower level operations
	hook := NewQueryInitStatusHook(i)
	ctx := statushooks.AddStatusHooksToContext(parentCtx, hook)

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
		i.Result.Error = fmt.Errorf("failed to load workspace: %s", errAndWarnings.GetError().Error())
		return
	}
	i.Result.AddWarnings(errAndWarnings.Warnings...)
	i.Workspace = w

	statushooks.SetStatus(ctx, "Resolving arguments")

	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	resolvedQueries, preparedStatementSource, err := w.GetQueriesFromArgs(args)
	if err != nil {
		i.Result.Error = err
		return
	}
	// create a cancellable context so that we can cancel the initialisation
	ctx, cancel := context.WithCancel(ctx)
	// and store it
	i.cancelInitialisation = cancel
	i.Queries = resolvedQueries
	i.PreparedStatementSource = preparedStatementSource

	// and call base init
	i.InitData.Init(ctx, constants.InvokerQuery)

}
