package query

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v4/telemetry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/workspace"
	"log"
)

// TODO KAI combine with initialisation.InitData
type InitData struct {
	Workspace         *workspace.Workspace
	Client            db_common.Client
	Result            *db_common.InitResult
	Loaded            chan struct{}
	ShutdownTelemetry func()
	Queries           []string
	cancel            context.CancelFunc
}

// NewInitData creates a new InitData object and returns it
// it also starts an asynchronous population of the object
// InitData.Done closes after asynchronous initialization completes
func NewInitData(ctx context.Context, w *workspace.Workspace, args []string) *InitData {
	i := new(InitData)

	i.Result = new(db_common.InitResult)
	i.Loaded = make(chan struct{})

	go i.init(ctx, w, args)

	return i
}

func (i *InitData) Cancel() {
	// cancel any ongoing operation
	if i.cancel != nil {
		i.cancel()
	}
	i.cancel = nil
}

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
		if r := recover(); r != nil {
			i.Result.Error = helpers.ToError(r)
		}
		if i.Result.Error == nil {
			i.Result.Error = ctx.Err()
		}
		i.cancel = nil
		// close loaded channel to indicate we are complete
		close(i.Loaded)
	}()
	// create a cancellable context so that we can cancel the initialisation
	ctx, cancel := context.WithCancel(ctx)
	// and store it
	i.cancel = cancel

	// init telemetry
	shutdownTelemetry, err := telemetry.Init(constants.AppName)
	if err != nil {
		i.Result.AddWarnings(err.Error())
	} else {
		i.ShutdownTelemetry = shutdownTelemetry
	}

	// check if the required plugins are installed
	if err := w.CheckRequiredPluginsInstalled(); err != nil {
		i.Result.Error = err
		return
	}
	i.Workspace = w

	//validate steampipe version
	if err = w.ValidateSteampipeVersion(); err != nil {
		i.Result.Error = err
		return
	}

	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	queries, preparedStatementSource, err := w.GetQueriesFromArgs(args)
	if err != nil {
		i.Result.Error = err
		return
	}
	i.Queries = queries

	// set up the session data - prepared statements and introspection tables
	// this defaults to creating prepared statements for all queries
	sessionDataSource := workspace.NewSessionDataSource(w, preparedStatementSource)

	// get the client
	// set max DB connections to 1
	viper.Set(constants.ArgMaxParallel, 1)

	// TODO KAI is FDW message working in interactive - does it rerender prompt??

	// add a message rendering function to the context - this is used for the fdw update message and
	// allows us to render it as a standard initialisation message
	getClientCtx := statushooks.AddMessageRendererToContext(ctx, func(format string, a ...any) {
		i.Result.AddMessage(fmt.Sprintf(format, a...))
	})

	// define db connection callback function
	ensureSessionData := func(ctx context.Context, conn *pgx.Conn) error {
		err, warnings := workspace.EnsureSessionData(ctx, sessionDataSource, conn)
		// TODO KAI how do we display wanrings
		log.Println("[WARN]", warnings)
		return err
	}

	c, err := initialisation.GetDbClient(getClientCtx, constants.InvokerQuery, ensureSessionData)
	if err != nil {
		i.Result.Error = err
		return
	}
	i.Client = c

	res := i.Client.RefreshConnectionAndSearchPaths(ctx)
	if res.Error != nil {
		i.Result.Error = res.Error
		return
	}
	i.Result.AddWarnings(res.Warnings...)

	// TODO KAI ensure we display prepared statement warnings
}
