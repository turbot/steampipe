package query

import (
	"context"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/workspace"
)

// TODO KAI combine with initialisation.InitData
type InitData struct {
	cancelInitialisation context.CancelFunc
	initialisation.InitData
	Loaded chan struct{}
	// map of query name to query (key is the query text for command line queries)
	Queries map[string]string
}

// NewInitData creates a new InitData object and returns it
// it also starts an asynchronous population of the object
// InitData.Done closes after asynchronous initialization completes
func NewInitData(ctx context.Context, w *workspace.Workspace, args []string) *InitData {
	i := &InitData{
		InitData: initialisation.InitData{
			Workspace: w,
			Result:    &db_common.InitResult{},
		},
		Loaded: make(chan struct{}),
	}

	go i.init(ctx, w, args)

	return i
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

func (i *InitData) init(ctx context.Context, w *workspace.Workspace, args []string) {
	defer func() {
		close(i.Loaded)
		// clear the cancelInitialisation function
		i.cancelInitialisation = nil
	}()
	// set max DB connections to 1
	viper.Set(constants.ArgMaxParallel, 1)
	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	queries, preparedStatementSource, err := w.GetQueriesFromArgs(args)
	if err != nil {
		i.Result.Error = err
		return
	}
	// create a cancellable context so that we can cancel the initialisation
	ctx, cancel := context.WithCancel(ctx)
	// and store it
	i.cancelInitialisation = cancel
	i.Queries = queries
	i.PreparedStatementSource = preparedStatementSource
	// now disable status hooks and call base init
	ctx = statushooks.DisableStatusHooks(ctx)
	i.InitData.Init(ctx, constants.InvokerQuery)
}
