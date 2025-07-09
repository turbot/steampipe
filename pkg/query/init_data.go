package query

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_client"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/export"
	"github.com/turbot/steampipe/v2/pkg/initialisation"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
)

type InitData struct {
	initialisation.InitData

	cancelInitialisation context.CancelFunc
	StartTime            time.Time
	Loaded               chan struct{}
	// map of query name to resolved query (key is the query text for command line queries)
	Queries []*modconfig.ResolvedQuery
}

// NewInitData returns a new InitData object
// It also starts an asynchronous population of the object
// InitData.Done closes after asynchronous initialization completes
func NewInitData(ctx context.Context, args []string) *InitData {
	i := &InitData{
		StartTime: time.Now(),
		InitData:  *initialisation.NewInitData(),
		Loaded:    make(chan struct{}),
	}

	statushooks.SetStatus(ctx, "Loading workspace")

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
	if len(viper.GetStringSlice(pconstants.ArgExport)) > 0 {
		i.RegisterExporters(queryExporters()...)

		// validate required export formats
		if err := i.ExportManager.ValidateExportFormat(viper.GetStringSlice(pconstants.ArgExport)); err != nil {
			i.Result.Error = err
			return
		}
	}

	// set max DB connections to 1
	viper.Set(pconstants.ArgMaxParallel, 1)

	statushooks.SetStatus(ctx, "Resolving arguments")

	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	resolvedQueries, err := getQueriesFromArgs(args)
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

// getQueriesFromArgs retrieves queries from args
//
// For each arg check if it is a named query or a file, before falling back to treating it as sql
func getQueriesFromArgs(args []string) ([]*modconfig.ResolvedQuery, error) {

	var queries = make([]*modconfig.ResolvedQuery, len(args))
	for idx, arg := range args {
		resolvedQuery, err := ResolveQueryAndArgsFromSQLString(arg)
		if err != nil {
			return nil, err
		}
		if len(resolvedQuery.ExecuteSQL) > 0 {
			// default name to the query text
			resolvedQuery.Name = resolvedQuery.ExecuteSQL

			queries[idx] = resolvedQuery
		}
	}
	return queries, nil
}

// ResolveQueryAndArgsFromSQLString attempts to resolve 'arg' to a query and query args
func ResolveQueryAndArgsFromSQLString(sqlString string) (*modconfig.ResolvedQuery, error) {
	var err error

	// 2) is this a file
	// get absolute filename
	filePath, err := filepath.Abs(sqlString)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}
	fileQuery, fileExists, err := getQueryFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}
	if fileExists {
		if fileQuery.ExecuteSQL == "" {
			error_helpers.ShowWarning(fmt.Sprintf("file '%s' does not contain any data", filePath))
			// (just return the empty query - it will be filtered above)
		}
		return fileQuery, nil
	}
	// the argument cannot be resolved as an existing file
	// if it has a sql suffix (i.e we believe the user meant to specify a file) return a file not found error
	if strings.HasSuffix(strings.ToLower(sqlString), ".sql") {
		return nil, fmt.Errorf("file '%s' does not exist", filePath)
	}

	// 2) just use the query string as is and assume it is valid SQL
	return &modconfig.ResolvedQuery{RawSQL: sqlString, ExecuteSQL: sqlString}, nil
}

// try to treat the input string as a file name and if it exists, return its contents
func getQueryFromFile(input string) (*modconfig.ResolvedQuery, bool, error) {
	// get absolute filename
	path, err := filepath.Abs(input)
	if err != nil {
		//nolint:golint,nilerr // if this gives any error, return not exist
		return nil, false, nil
	}

	// does it exist?
	if _, err := os.Stat(path); err != nil {
		//nolint:golint,nilerr // if this gives any error, return not exist (we may get a not found or a path too long for example)
		return nil, false, nil
	}

	// read file
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, true, err
	}

	res := &modconfig.ResolvedQuery{
		RawSQL:     string(fileBytes),
		ExecuteSQL: string(fileBytes),
	}
	return res, true, nil
}
