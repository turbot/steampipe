package query

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_client"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

type InitData struct {
	Loaded    chan struct{}
	Queries   []string
	Workspace *workspace.Workspace
	Client    db_common.Client
	Result    *db_common.InitResult
	cancel    context.CancelFunc
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

func (i *InitData) Cleanup(ctx context.Context) {
	// cancel any ongoing operation
	if i.cancel != nil {
		i.cancel()
	}

	// ensure that the initialisation was completed
	// and that we are not in a race condition where
	// the client is set after the cancel hits
	<-i.Loaded

	// if a client was initialised, close it
	if i.Client != nil {
		i.Client.Close(ctx)
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
		close(i.Loaded)
	}()

	// create a cancellable context so that we can cancel the initialisation
	ctx, cancel := context.WithCancel(ctx)
	// and store it
	i.cancel = cancel

	// set max DB connections to 1
	viper.Set(constants.ArgMaxParallel, 1)

	c, err := getClient(ctx)
	if err != nil {
		i.Result.Error = err
		return
	}
	i.Client = c

	// check if the required plugins are installed
	if err := w.CheckRequiredPluginsInstalled(); err != nil {
		i.Result.Error = err
		return
	}
	i.Workspace = w

	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	queries, preparedStatementSource, err := w.GetQueriesFromArgs(args)
	if err != nil {
		i.Result.Error = err
		return
	}
	i.Queries = queries

	res := i.Client.RefreshConnectionAndSearchPaths(ctx)
	if res.Error != nil {
		i.Result.Error = res.Error
		return
	}
	i.Result.AddWarnings(res.Warnings...)

	// set up the session data - prepared statements and introspection tables
	// this defaults to creating prepared statements for all queries
	sessionDataSource := workspace.NewSessionDataSource(w, preparedStatementSource)

	// register EnsureSessionData as a callback on the client.
	// if the underlying SQL client has certain errors (for example context expiry) it will reset the session
	// so our client object calls this callback to restore the session data
	i.Client.SetEnsureSessionDataFunc(func(ctx context.Context, session *db_common.DatabaseSession) (error, []string) {
		return workspace.EnsureSessionData(ctx, sessionDataSource, session)
	})

	// force creation of session data - se we see any prepared statement errors at once
	sessionResult := i.Client.AcquireSession(ctx)
	i.Result.AddWarnings(sessionResult.Warnings...)
	if sessionResult.Error != nil {
		i.Result.Error = fmt.Errorf("error acquiring database connection, %s", sessionResult.Error.Error())
	} else {
		sessionResult.Session.Close(utils.IsContextCancelled(ctx))
	}
}

func getClient(ctx context.Context) (db_common.Client, error) {
	var client db_common.Client
	var err error
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		client, err = db_client.NewDbClient(ctx, connectionString)
	} else {
		client, err = db_local.GetLocalClient(ctx, constants.InvokerQuery)
	}
	return client, err
}
