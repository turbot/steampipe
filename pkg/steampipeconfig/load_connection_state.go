package steampipeconfig

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
)

// ConnectionStateTableAddedColumns is a map of column names to the SQL type
// these are columns which needed to be added after 0.20.0 with the
// steampipe_connection_state table was released
var ConnectionStateTableAddedColumns map[string]string = map[string]string{
	"connections": "TEXT[]",
}

// LoadConnectionState populates a ConnectionStateMap from the connection_state table
// it verifies the table has been initialised by calling RefreshConnections after db startup
func LoadConnectionState(ctx context.Context, conn *pgx.Conn, opts ...LoadConnectionStateOption) (ConnectionStateMap, error) {
	config := &LoadConnectionStateConfiguration{}
	for _, opt := range opts {
		opt(config)
	}

	// max duration depends on if waiting for ready or just pending
	// default value is if we are waiting for pending
	// set this to a long enough time for ConnectionUpdates to be generated for a large connection count
	// TODO this time can be reduced once all; plugins are using v5.4.1 of the sdk
	maxDuration := 1 * time.Minute
	retryInterval := 250 * time.Millisecond
	if config.WaitMode == WaitForReady || config.WaitMode == WaitForSearchPath {
		// is we are waiting for all connections to be ready, wait up to 10 minutes
		maxDuration = 10 * time.Minute
	}
	backoff := retry.NewConstant(retryInterval)

	var connectionStateMap ConnectionStateMap

	err := retry.Do(ctx, retry.WithMaxDuration(maxDuration, backoff), func(ctx context.Context) error {
		var loadErr error
		connectionStateMap, loadErr = loadConnectionState(ctx, conn)
		if loadErr != nil {
			return loadErr
		}

		// now process any load options
		switch config.WaitMode {
		case WaitForReady:
			return checkConnectionsAreReady(ctx, connectionStateMap, config)
		case WaitForLoading:
			if connectionStateMap.Pending() {
				return retry.RetryableError(fmt.Errorf("timed out waiting for connection state to be updated from pending"))
			}
		case WaitForSearchPath:
			if len(config.SearchPath) == 0 {
				// nothing to do
				return nil
			}

			// wait for search path is called with a search path set - we must convert this into a set of
			// connections which we must wait for (the first connection for each plugin)
			// the first time we load the connection state, determine the connections we need to wait for
			if len(config.Connections) == 0 {
				// build list of connections we must wait for as update config
				config.Connections = connectionStateMap.GetFirstSearchPathConnectionForPlugins(config.SearchPath)
			}
			// now check if these connections are ready
			if err := checkConnectionsAreReady(ctx, connectionStateMap, config); err != nil {
				return err
			}

			// so all required connections are loaded, either 'ready' or 'error'
			// verify that not all schemas are in error state
			// (this returns an error if any schemas are in error state)
			if allConnectionsInError(config.Connections, connectionStateMap) {
				return fmt.Errorf("all connections in search path are in error")
			}
			return nil

		}
		return nil

	})

	return connectionStateMap, err
}

func loadConnectionState(ctx context.Context, conn *pgx.Conn, opts ...loadConnectionStateOption) (ConnectionStateMap, error) {
	config := &loadConnectionStateConfig{}
	for _, configOption := range opts {
		configOption(config)
	}
	log.Println("[TRACE] with config", config)

	query := buildLoadConnectionStateQuery(config)

	log.Println("[TRACE] running query", query)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		// columns were added after the 0.20.0 release (connections for now)
		// we need to handle the case where we are connected to an old version of
		// service which doesn't have some of these columns
		if column, isColumNotFound := isColumnNotFoundError(err); isColumNotFound {
			// if this was not an added column, return the error as is
			if _, isAddedColumn := ConnectionStateTableAddedColumns[column]; !isAddedColumn {
				return nil, err
			}
			// try to load with the added column ignored
			return loadConnectionState(ctx, conn, ignoreColumns(append(config.ignoredColumns, column)...))
		}
		return nil, err
	}
	defer rows.Close()

	var res = make(ConnectionStateMap)

	connectionStateList, err := pgx.CollectRows(rows, pgx.RowToStructByName[ConnectionState])
	if err != nil {
		return nil, err
	}

	for _, c := range connectionStateList {
		// copy into loop var
		connectionState := c
		res[c.ConnectionName] = &connectionState
	}

	return res, nil
}

func checkConnectionsAreReady(ctx context.Context, connectionStateMap ConnectionStateMap, config *LoadConnectionStateConfiguration) error {
	if !connectionStateMap.Loaded(config.Connections...) {
		statusMessage := GetLoadingConnectionStatusMessage(connectionStateMap, config.Connections...)
		statushooks.SetStatus(ctx, statusMessage)
		return retry.RetryableError(fmt.Errorf("connection state is still loading"))
	}
	return nil
}

func allConnectionsInError(connectionsNames []string, connectionStateMap ConnectionStateMap) bool {
	if len(connectionsNames) == 0 {
		return false
	}
	for _, connectionName := range connectionsNames {
		connectionState, ok := connectionStateMap[connectionName]
		if !ok {
			// not expected but not impossible - state may have changed while we iterate
			continue
		}
		if connectionState.State != constants.ConnectionStateError {
			return false
		}
	}

	return true
}

func GetLoadingConnectionStatusMessage(connectionStateMap ConnectionStateMap, requiredSchemas ...string) string {
	var connectionSummary = connectionStateMap.GetSummary()

	readyCount := connectionSummary[constants.ConnectionStateReady]
	totalCount := len(connectionStateMap) - connectionSummary[constants.ConnectionStateDeleting]

	loadedMessage := fmt.Sprintf("Loaded %d of %d %s",
		readyCount,
		totalCount,
		utils.Pluralize("connection", totalCount))

	if len(requiredSchemas) == 1 {
		// if we are only waiting for a single schema, include that in the message
		return fmt.Sprintf("Waiting for connection '%s' to load (%s)", requiredSchemas[0], loadedMessage)
	}

	return loadedMessage
}

func SaveConnectionStateFile(res *RefreshConnectionResult, connectionUpdates *ConnectionUpdates) {
	// now serialise the connection state
	connectionState := make(ConnectionStateMap, len(connectionUpdates.FinalConnectionState))
	for k, v := range connectionUpdates.FinalConnectionState {
		connectionState[k] = v
	}
	// NOTE: add any connection which failed
	for c, reason := range res.FailedConnections {
		connectionState[c].State = constants.ConnectionStateError
		connectionState[c].SetError(reason)
	}

	// update connection state and write the missing and failed plugin connections
	if err := connectionState.Save(); err != nil {
		res.Error = err
	}
}

func DeleteConnectionStateFile() {
	os.Remove(filepaths.ConnectionStatePath())
}

func isColumnNotFoundError(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	pgErr, ok := err.(*pgconn.PgError)
	if !ok || pgErr.Code != "42703" {
		return "", false
	}

	// at this point, we know that it's a PgError with code 42703 (column not found)
	// let's try to find out the name of the column

	// try to find out the name from <tablename>.<columnname>
	r := regexp.MustCompile(`^column "(.*)\.(.*)" does not exist$`)
	captureGroups := r.FindStringSubmatch(pgErr.Message)
	if len(captureGroups) == 3 {
		return captureGroups[2], true
	}

	// maybe there is no table name
	// try to find out the name from <columnname>
	r = regexp.MustCompile(`^column "(.*)" does not exist$`)
	captureGroups = r.FindStringSubmatch(pgErr.Message)
	if len(captureGroups) == 2 {
		return captureGroups[1], true
	}
	return "", true
}

// buildLoadConnectionStateQuery builds up the SQL we send to the service
func buildLoadConnectionStateQuery(config *loadConnectionStateConfig) string {
	prefix := `SELECT name,
	type,
	import_schema,
	state,
	error,	
	plugin,
	schema_mode,
	schema_hash,
	comments_set,
	connection_mod_time,
	plugin_mod_time`

	// because columns were added post 0.20.0 release, we have to handle cases
	// where we are selecting from a service which doesn't have the columns added later
	//
	// for every colmn added
	//		is it ignored already -> select "NULL as colname"
	//		else select the column value
	var extraCols []string
	ignoreLookup := utils.SliceToLookup(config.ignoredColumns)
	for ignoreColumn := range ConnectionStateTableAddedColumns {
		// is this ignored
		if _, ignored := ignoreLookup[ignoreColumn]; ignored {
			// read NULL for this column
			extraCols = append(extraCols, fmt.Sprintf("NULL as %s", ignoreColumn))
		} else {
			extraCols = append(extraCols, ignoreColumn)
		}
	}

	query := fmt.Sprintf(
		`%s,%s FROM %s.%s `,
		prefix,
		strings.Join(extraCols, ",\n"),
		constants.InternalSchema,
		constants.ConnectionStateTable,
	)
	return query
}

type loadConnectionStateConfig struct {
	ignoredColumns []string
}

type loadConnectionStateOption func(l *loadConnectionStateConfig)

func ignoreColumns(cols ...string) loadConnectionStateOption {
	return func(l *loadConnectionStateConfig) {
		l.ignoredColumns = cols
	}
}
