package connection

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/connection_state"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/semaphore"
)

type connectionError struct {
	name string
	err  error
}

type refreshConnectionState struct {
	pool                       *pgxpool.Pool
	searchPath                 []string
	connectionUpdates          *steampipeconfig.ConnectionUpdates
	tableUpdater               *connectionStateTableUpdater
	res                        *steampipeconfig.RefreshConnectionResult
	forceUpdateConnectionNames []string

	// properties for schema/comment cloning
	exemplarSchemaMapMut sync.Mutex

	// maps keyed by plugin which gives an exemplar connection name,
	// if a plugin has an entry in this map, all connections schemas can be cloned from teh exemplar schema
	exemplarSchemaMap map[string]string
	// if a plugin has an entry in this map, all connections schemas can be cloned from teh exemplar schema
	exemplarCommentsMap map[string]string
}

func newRefreshConnectionState(ctx context.Context, forceUpdateConnectionNames []string) (*refreshConnectionState, error) {
	// create a connection pool to connection refresh
	poolsize := 20
	pool, err := db_local.CreateConnectionPool(ctx, &db_local.CreateDbOptions{Username: constants.DatabaseSuperUser}, poolsize)
	if err != nil {
		return nil, err
	}

	// set user search path first
	log.Printf("[INFO] setting up search path")
	searchPath, err := db_local.SetUserSearchPath(ctx, pool)
	if err != nil {
		// note: close pool in case of error
		pool.Close()
		return nil, err
	}

	return &refreshConnectionState{
		pool:                       pool,
		searchPath:                 searchPath,
		forceUpdateConnectionNames: forceUpdateConnectionNames,
	}, nil
}

func (state *refreshConnectionState) close() {
	if state.pool != nil {
		state.pool.Close()
	}
}

// RefreshConnections loads required connections from config
// and update the database schema and search path to reflect the required connections
// return whether any changes have been made
func (state *refreshConnectionState) refreshConnections(ctx context.Context) {

	// if there was an error (other than a connection error, which will NOT have been assigned to res),
	// set state of all incomplete connections to error
	defer func() {
		if state.res != nil && state.res.Error != nil {
			state.setIncompleteConnectionStateToError(ctx, sperr.WrapWithMessage(state.res.Error, "refreshConnections failed before connection update was complete"))
			// TODO send error PG notification
		}
	}()
	log.Printf("[INFO] building connectionUpdates")

	// determine any necessary connection updates
	state.connectionUpdates, state.res = steampipeconfig.NewConnectionUpdates(ctx, state.pool, state.forceUpdateConnectionNames...)
	defer state.logRefreshConnectionResults()
	// were we successful#
	if state.res.Error != nil {
		return
	}

	log.Printf("[INFO] created connection updates")

	// delete the connection state file - it will be rewritten when we are complete
	log.Printf("[INFO] deleting connections state file")
	steampipeconfig.DeleteConnectionStateFile()
	defer func() {
		if state.res.Error == nil {
			log.Printf("[INFO] saving connections state file")
			steampipeconfig.SaveConnectionStateFile(state.res, state.connectionUpdates)
		}
	}()

	// warn about missing plugins
	state.addMissingPluginWarnings()

	// create object to update the connection state table and notify of state changes
	state.tableUpdater = newConnectionStateTableUpdater(state.connectionUpdates, state.pool)

	// update connectionState table to reflect the updates (i.e. set connections to updating/deleting/ready as appropriate)
	// also this will update the schema hashes of plugins
	if err := state.tableUpdater.start(ctx); err != nil {
		state.res.Error = err
		return
	}

	// if there are no updates, just return
	if !state.connectionUpdates.HasUpdates() {
		log.Println("[INFO] no updates required")
		return
	}

	log.Printf("[INFO] execute connection queries")

	// execute any necessary queries
	state.executeConnectionQueries(ctx)
	if state.res.Error != nil {
		log.Printf("[WARN] refreshConnections failed with err %s", state.res.Error.Error())
		return
	}

	state.res.UpdatedConnections = true
}

func (state *refreshConnectionState) addMissingPluginWarnings() {
	log.Printf("[INFO] refreshConnections: identify missing plugins")

	var connectionNames, pluginNames []string
	// add warning if there are connections left over, from missing plugins
	if len(state.connectionUpdates.MissingPlugins) > 0 {
		// warning
		for a, conns := range state.connectionUpdates.MissingPlugins {
			for _, con := range conns {
				connectionNames = append(connectionNames, con.Name)
			}
			pluginNames = append(pluginNames, utils.GetPluginName(a))
		}
		state.res.AddWarning(fmt.Sprintf("%d %s required by %s %s missing. To install, please run %s",
			len(pluginNames),
			utils.Pluralize("plugin", len(pluginNames)),
			utils.Pluralize("connection", len(connectionNames)),
			utils.Pluralize("is", len(pluginNames)),
			constants.Bold(fmt.Sprintf("steampipe plugin install %s", strings.Join(pluginNames, " ")))))
	}
}

func (state *refreshConnectionState) logRefreshConnectionResults() {
	var cmdName = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command).Name()
	if cmdName != "plugin-manager" {
		return
	}

	var op strings.Builder
	if state.connectionUpdates != nil {
		op.WriteString(fmt.Sprintf("%s", state.connectionUpdates.String()))
	}
	if state.res != nil {
		op.WriteString(fmt.Sprintf("%s\n", state.res.String()))
	}

	log.Printf("[TRACE] refresh connections: \n%s\n", helpers.Tabify(op.String(), "    "))
}

func (state *refreshConnectionState) executeConnectionQueries(ctx context.Context) {
	// retrieve updates from the table updater
	connectionUpdates := state.tableUpdater.updates

	utils.LogTime("db.executeConnectionQueries start")
	defer utils.LogTime("db.executeConnectionQueries start")

	// execute deletions
	if err := state.executeDeleteQueries(ctx); err != nil {
		// just log
		log.Printf("[WARN] failed to delete all unused schemas: %s", err.Error())
	}

	// execute updates
	numUpdates := len(connectionUpdates.Update)
	numMissingComments := len(connectionUpdates.MissingComments)
	log.Printf("[INFO] executeConnectionQueries: num updates: %d, connections missing comments: %d", numUpdates, numMissingComments)

	if numUpdates+numMissingComments > 0 {
		// get schema queries - this updates schemas for validated plugins and drops schemas for unvalidated plugins
		state.executeUpdateQueries(ctx)
	} else if len(connectionUpdates.Delete) > 0 {
		log.Printf("[INFO] deleted all unnecessary schemas - sending notification")

		// if there are no updates and there ARE deletes, notify
		// (is there are updates, deletes will be notified by executeUpdateQueries)
		if err := state.sendPostgreSchemaNotification(ctx); err != nil {
			// just log
			log.Printf("[WARN] failed to send schema deletion Postgres notification: %s", err.Error())
		}

	}

	return
}

// execute all update queries
// NOTE: this only sets res.Error if there is a failure to set update the connection state table
// - all other connection based failures are recorded in the connection state table
func (state *refreshConnectionState) executeUpdateQueries(ctx context.Context) {
	utils.LogTime("db.executeUpdateQueries start")
	defer utils.LogTime("db.executeUpdateQueries end")

	defer func() {
		if state.res.Error != nil {
			log.Printf("[INFO] executeUpdateQueries returned error: %v", state.res.Error)
		}
	}()

	connectionUpdates := state.connectionUpdates
	connectionPlugins := connectionUpdates.ConnectionPlugins
	numUpdates := len(connectionUpdates.Update)

	// we need to execute the updates in search path order
	// i.e. we first need to update the first search path connection for each plugin (this can be done in parallel)
	// then we can update the remaining connections in parallel
	initialUpdates, remainingUpdates, dynamicUpdates := state.getInitialAndRemainingUpdates()

	// dynamic plugins must be updated for each plugin in search path order
	// dynamicUpdates is a map keyed by plugin with all the updates for that plugin

	// create exemplar maps
	state.exemplarSchemaMap = make(map[string]string)
	state.exemplarCommentsMap = make(map[string]string)
	log.Printf("[INFO] executing %d update %s", numUpdates, utils.Pluralize("query", numUpdates))

	// execute initial updates
	log.Printf("[INFO] executing initial updates")
	var errors []error
	moreErrors := state.executeUpdatesInParallel(ctx, initialUpdates)
	errors = append(errors, moreErrors...)

	// execute dynamic updates (not, we update all connections in search path order,
	// so must call executeUpdateSetsInParallel)
	log.Printf("[INFO] executing dynamic updates")
	moreErrors = state.executeUpdateSetsInParallel(ctx, dynamicUpdates)
	errors = append(errors, moreErrors...)

	// if any of the initial schemas failed, do not proceed - these schemas are required to ensure we correctly
	// resolve unqualified queries/tables
	if len(errors) > 0 {
		state.res.Error = error_helpers.CombineErrors(errors...)
		log.Printf("[WARN] initial updates failed: %s", state.res.Error.Error())
		// TODO SEND ERROR NOTIFICATION
		return
	}

	log.Printf("[INFO] set comments for initial updates")
	// now set comments for initial updates and dynamic connections
	// note errors will be empty to get here
	state.UpdateCommentsInParallel(ctx, maps.Values(initialUpdates), connectionPlugins)

	log.Printf("[INFO] set comments for dynamic updates")
	// convert dynamicUpdates to an array of connection states
	var dynamicUpdateArray = updateSetMapToArray(dynamicUpdates)
	state.UpdateCommentsInParallel(ctx, dynamicUpdateArray, connectionPlugins)

	log.Printf("[INFO] updated all exemplar schemas - sending notification")
	// now that we have updated all exemplar schemars, send postgres notification
	// this gives any attached interactive clients a chance to update their inspect data and autocomplete

	if err := state.sendPostgreSchemaNotification(ctx); err != nil {
		// just log
		log.Printf("[WARN] failed to send schem update Postgres notification: %s", err.Error())
	}

	log.Printf("[INFO] Execute %d remaining %s",
		len(remainingUpdates),
		utils.Pluralize("updates", len(remainingUpdates)))
	// now execute remaining updates
	moreErrors = state.executeUpdatesInParallel(ctx, remainingUpdates)
	errors = append(errors, moreErrors...)

	log.Printf("[INFO] Set comments for %d remaining %s and %d %s missing comments",
		len(remainingUpdates),
		utils.Pluralize("updates", len(remainingUpdates)),
		len(connectionUpdates.MissingComments),
		utils.Pluralize("updates", len(connectionUpdates.MissingComments)),
	)
	// set comments for remaining updates
	state.UpdateCommentsInParallel(ctx, maps.Values(remainingUpdates), connectionPlugins)
	// set comments for any other connection without comment set
	state.UpdateCommentsInParallel(ctx, maps.Values(state.connectionUpdates.MissingComments), connectionPlugins)

	if len(errors) > 0 {
		state.res.Error = error_helpers.CombineErrors(errors...)
	}

	log.Printf("[INFO] all update queries executed")

	for _, failure := range connectionUpdates.InvalidConnections {
		log.Printf("[TRACE] remove schema for connection failing validation connection %s, plugin Name %s\n ", failure.ConnectionName, failure.Plugin)
		if failure.ShouldDropIfExists {
			_, err := state.pool.Exec(ctx, db_common.GetDeleteConnectionQuery(failure.ConnectionName))
			if err != nil {
				// NOTE: do not return an error if we fail to remove an invalid connection - just log it
				log.Printf("[WARN] failed to delete invalid connection '%s' (%s) : %s", failure.ConnectionName, failure.Message, err.Error())
			}
		}
	}
	log.Printf("[INFO] executeUpdateQueries complete")
	return
}

// convert map upd update sets (used for dynamic schemas) to an array of the underlying connection states
func updateSetMapToArray(updateSetMap map[string][]*steampipeconfig.ConnectionState) []*steampipeconfig.ConnectionState {
	var res []*steampipeconfig.ConnectionState
	for _, updates := range updateSetMap {
		res = append(res, updates...)
	}
	return res
}

// create/update connections

func (state *refreshConnectionState) executeUpdatesInParallel(ctx context.Context, updates map[string]*steampipeconfig.ConnectionState) (errors []error) {
	// just call executeUpdateSetsInParallel

	// convert updates to update sets
	updatesAsSets := make(map[string][]*steampipeconfig.ConnectionState, len(updates))
	for k, v := range updates {
		updatesAsSets[k] = []*steampipeconfig.ConnectionState{v}
	}
	return state.executeUpdateSetsInParallel(ctx, updatesAsSets)
}

// execute sets of updates in parallel - this is required as for dynamic plugins, we must updated all connections in
// search path order
// - for convenience we also use this function for static connections by mapping the input data
// from map[string]*steampipeconfig.ConnectionState to map[string][]*steampipeconfig.ConnectionState
func (state *refreshConnectionState) executeUpdateSetsInParallel(ctx context.Context, updates map[string][]*steampipeconfig.ConnectionState) (errors []error) {
	var wg sync.WaitGroup
	var errChan = make(chan *connectionError)

	// use as many goroutines as we have connections
	var maxUpdateThreads = int64(state.pool.Config().MaxConns)
	sem := semaphore.NewWeighted(maxUpdateThreads)

	go func() {
		for {
			select {
			case connectionError := <-errChan:
				if connectionError == nil {
					return
				}
				errors = append(errors, connectionError.err)
				conn, poolErr := state.pool.Acquire(ctx)
				if poolErr == nil {
					state.tableUpdater.onConnectionError(ctx, conn.Conn(), connectionError.name, connectionError.err)
					conn.Release()
				}
			}
		}
	}()

	// each update may be multiple connections, to execute in order
	for _, states := range updates {
		wg.Add(1)
		// use semaphore to limit goroutines
		if err := sem.Acquire(ctx, 1); err != nil {
			errors = append(errors, err)
			// if we fail to acquire semaphore, just give up
			return errors
		}
		go func(connectionStates []*steampipeconfig.ConnectionState) {
			defer func() {
				wg.Done()
				sem.Release(1)
			}()

			state.executeUpdateForConnections(ctx, errChan, connectionStates...)
		}(states)

	}

	wg.Wait()
	close(errChan)

	return errors
}

// syncronously execute the update queries for one or more connections
func (state *refreshConnectionState) executeUpdateForConnections(ctx context.Context, errChan chan *connectionError, connectionStates ...*steampipeconfig.ConnectionState) {
	for _, connectionState := range connectionStates {
		connectionName := connectionState.ConnectionName
		remoteSchema := utils.PluginFQNToSchemaName(connectionState.Plugin)
		var sql string

		state.exemplarSchemaMapMut.Lock()
		// is this plugin in the exemplarSchemaMap
		exemplarSchemaName, haveExemplarSchema := state.exemplarSchemaMap[connectionState.Plugin]
		if haveExemplarSchema {
			// we can clone!
			sql = getCloneSchemaQuery(sql, exemplarSchemaName, connectionState)
		} else {
			// just get sql to execute update query, and update the connection state table, in a transaction
			sql = db_common.GetUpdateConnectionQuery(connectionName, remoteSchema)
		}
		state.exemplarSchemaMapMut.Unlock()

		// the only error this will return is the failure to update the state table
		// - all other errors are written to the state table
		if err := state.executeUpdateQuery(ctx, sql, connectionName); err != nil {
			errChan <- &connectionError{connectionName, err}
		} else {
			// we can clone this plugin, add to exemplarSchemaMap
			// (AFTER executing the update query)
			if !haveExemplarSchema && connectionState.CanCloneSchema() {
				state.exemplarSchemaMap[connectionState.Plugin] = connectionName
			}
		}
	}
}

func (state *refreshConnectionState) executeUpdateQuery(ctx context.Context, sql, connectionName string) error {
	// create a transaction
	tx, err := state.pool.Begin(ctx)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to create transaction to perform update query")
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	// execute update sql
	_, err = tx.Exec(ctx, sql)
	if err != nil {
		// update failed connections in result
		state.res.AddFailedConnection(connectionName, err.Error())

		// update the state table
		//(the transaction will be aborted - create a connection for the update)
		if conn, poolErr := state.pool.Acquire(ctx); poolErr == nil {
			if statusErr := state.tableUpdater.onConnectionError(ctx, conn.Conn(), connectionName, err); statusErr != nil {
				// NOTE: do not return the error - unless we failed to update the connection state table
				return error_helpers.CombineErrorsWithPrefix(fmt.Sprintf("failed to update connection %s and failed to update connection_state table", connectionName), err, statusErr)
			}
		}
		return nil
	}

	// update state table (inside transaction)
	err = state.tableUpdater.onConnectionReady(ctx, tx.Conn(), connectionName)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to update connection state table")
	}
	return nil
}

// set connection comments

func (state *refreshConnectionState) UpdateCommentsInParallel(ctx context.Context, updates []*steampipeconfig.ConnectionState, plugins map[string]*steampipeconfig.ConnectionPlugin) (errors []error) {
	if !viper.GetBool(constants.ArgSchemaComments) {
		return nil
	}

	var wg sync.WaitGroup
	var errChan = make(chan *connectionError)

	// use as many goroutines as we have connections
	var maxUpdateThreads = int64(state.pool.Config().MaxConns)
	sem := semaphore.NewWeighted(maxUpdateThreads)

	go func() {
		for {
			select {
			case connectionError := <-errChan:
				if connectionError == nil {
					return
				}
				errors = append(errors, connectionError.err)
				// TODO just log errors
			}
		}
	}()

	// each update may be multiple connections, to execute in order
	for _, connectionState := range updates {
		wg.Add(1)
		// use semaphore to limit goroutines
		if err := sem.Acquire(ctx, 1); err != nil {
			errors = append(errors, err)
			// if we fail to acquire semaphore, just give up
			return errors
		}
		go func(connectionState *steampipeconfig.ConnectionState) {
			defer func() {
				wg.Done()
				sem.Release(1)
			}()

			state.updateCommentsForConnection(ctx, errChan, plugins, connectionState)
		}(connectionState)

	}

	wg.Wait()
	close(errChan)

	return errors
}

// syncronously execute the comments queries for one or more connections
func (state *refreshConnectionState) updateCommentsForConnection(ctx context.Context, errChan chan *connectionError, connectionPluginMap map[string]*steampipeconfig.ConnectionPlugin, connectionState *steampipeconfig.ConnectionState) {
	connectionName := connectionState.ConnectionName

	var sql string

	// we should have a connectionPlugin loaded for this connection
	connectionPlugin, ok := connectionPluginMap[connectionName]
	if !ok {
		log.Printf("[WARN] no connection plugin loaded for connection '%s', which needs comments updating", connectionName)
		return
	}

	schema := connectionPlugin.ConnectionMap[connectionName].Schema.Schema
	// just get sql to execute update query, and update the connection state table, in a transaction
	sql = db_common.GetCommentsQueryForPlugin(connectionName, schema)

	// comment cloning disabled for now
	//// if this schema is static, add to the exemplar map
	//state.exemplarSchemaMapMut.Lock()
	//// is this plugin in the exemplarSchemaMap
	//exemplarSchemaName, haveExemplarSchema := state.exemplarCommentsMap[connectionState.Plugin]
	//if haveExemplarSchema {
	//// we can clone!
	//	sql = getCloneCommentsQuery(sql, exemplarSchemaName, connectionState)
	//} else {
	//	// get the schema from the connection plugin
	//	schema := connectionPluginMap[connectionName].ConnectionMap[connectionName].Schema.Schema
	//	// just get sql to execute update query, and update the connection state table, in a transaction
	//	sql = db_common.GetCommentsQueryForPlugin(connectionName, schema)
	//}
	//state.exemplarSchemaMapMut.Unlock()

	// the only error this will return is the failure to update the state table
	// - all other errors are written to the state table
	if err := state.executeCommentQuery(ctx, sql, connectionName); err != nil {
		errChan <- &connectionError{connectionName, err}
	} //else {
	//	// we can clone this plugin, add to exemplarCommentsMap
	//	// (AFTER executing the update query)
	//	if !haveExemplarSchema && connectionState.CanCloneSchema() {
	//		state.exemplarCommentsMap[connectionState.Plugin] = connectionName
	//	}
	//}
}

func (state *refreshConnectionState) executeCommentQuery(ctx context.Context, sql, connectionName string) error {
	// create a transaction
	tx, err := state.pool.Begin(ctx)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to create transaction to perform update query")
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	// execute update sql
	_, err = tx.Exec(ctx, sql)
	if err != nil {
		// update the state table
		//(the transaction will be aborted - create a connection for the update)
		if conn, poolErr := state.pool.Acquire(ctx); poolErr == nil {
			if statusErr := state.tableUpdater.onConnectionError(ctx, conn.Conn(), connectionName, err); statusErr != nil {
				// NOTE: do not return the error - unless we failed to update the connection state table
				return error_helpers.CombineErrorsWithPrefix(fmt.Sprintf("failed to update connection %s and failed to update connection_state table", connectionName), err, statusErr)
			}
		}

		return nil
	}

	// update state table (inside transaction)
	// ignore error
	if err := state.tableUpdater.onConnectionCommentsLoaded(ctx, tx.Conn(), connectionName); err != nil {
		log.Printf("[WARN] failed to set 'comments_set' for connection '%s': %s", connectionName, err.Error())
	}

	return nil
}

func getCloneSchemaQuery(sql string, exemplarSchemaName string, connectionState *steampipeconfig.ConnectionState) string {
	sql = fmt.Sprintf("select clone_foreign_schema('%s', '%s', '%s');", exemplarSchemaName, connectionState.ConnectionName, connectionState.Plugin)
	return sql
}

func getCloneCommentsQuery(sql string, exemplarSchemaName string, connectionState *steampipeconfig.ConnectionState) string {
	sql = fmt.Sprintf("select clone_table_comments('%s', '%s');", exemplarSchemaName, connectionState.ConnectionName)
	return sql
}

func (state *refreshConnectionState) getInitialAndRemainingUpdates() (initialUpdates, remainingUpdates map[string]*steampipeconfig.ConnectionState, dynamicUpdates map[string][]*steampipeconfig.ConnectionState) {
	updates := state.connectionUpdates.Update
	searchPathConnections := state.connectionUpdates.FinalConnectionState.GetFirstSearchPathConnectionForPlugins(state.searchPath)

	initialUpdates = make(map[string]*steampipeconfig.ConnectionState)
	remainingUpdates = make(map[string]*steampipeconfig.ConnectionState)
	// dynamic plugins must be updated for each plugin in search path order
	// build a map keyed by plugin, with the value the ordered updates for that plugin
	dynamicUpdates = make(map[string][]*steampipeconfig.ConnectionState)

	// convert this into a lookup of initial updates to execute
	for _, connectionName := range searchPathConnections {
		if connectionState, updateRequired := updates[connectionName]; updateRequired {
			if connectionState.SchemaMode == sdkplugin.SchemaModeDynamic {
				dynamicUpdates[connectionState.Plugin] = append(dynamicUpdates[connectionState.Plugin], connectionState)
			} else {
				initialUpdates[connectionName] = connectionState
			}
		}
	}
	// now add remaining updates to remainingUpdates
	for connectionName, connectionState := range updates {
		if _, isInitialUpdate := initialUpdates[connectionName]; !isInitialUpdate {
			remainingUpdates[connectionName] = connectionState
		}

	}
	return initialUpdates, remainingUpdates, dynamicUpdates
}

func (state *refreshConnectionState) executeDeleteQueries(ctx context.Context) error {
	deletions := maps.Keys(state.connectionUpdates.Delete)

	t := time.Now()
	log.Printf("[INFO] execute %d delete %s", len(deletions), utils.Pluralize("query", len(deletions)))
	defer func() {
		log.Printf("[INFO] completed execute delete queries (%fs)", time.Since(t).Seconds())
	}()

	var errors []error

	for _, c := range deletions {
		err := state.executeDeleteQuery(ctx, c)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return error_helpers.CombineErrors(errors...)
}

// delete the schema and update remove the connection from the state table
// NOTE: this only returns an error if we fail to update the state table
func (state *refreshConnectionState) executeDeleteQuery(ctx context.Context, connectionName string) error {
	// create a transaction
	tx, err := state.pool.Begin(ctx)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to create transaction to perform delete query")
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	sql := db_common.GetDeleteConnectionQuery(connectionName)

	// execute delete sql
	_, err = tx.Exec(ctx, sql)
	if err != nil {
		// update the state table
		//(the transaction will be aborted - create a connection for the update)
		if conn, poolErr := state.pool.Acquire(ctx); poolErr == nil {
			if statusErr := state.tableUpdater.onConnectionError(ctx, conn.Conn(), connectionName, err); statusErr != nil {
				// NOTE: do not return the error - unless we failed to update the connection state table
				return error_helpers.CombineErrorsWithPrefix(fmt.Sprintf("failed to update connection %s and failed to update connection_state table", connectionName), err, statusErr)
			}
		}

		return nil
	}

	// delete state table entry (inside transaction)
	err = state.tableUpdater.onConnectionDeleted(ctx, tx.Conn(), connectionName)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to delete connection state table entry for '%s'", connectionName)
	}
	return nil
}

// set the state of any incomplete connections to error
func (state *refreshConnectionState) setIncompleteConnectionStateToError(ctx context.Context, err error) {
	// create wrapped error
	connectionStateError := sperr.WrapWithMessage(err, "failed to update Steampipe connections")
	// load connection state
	conn, err := state.pool.Acquire(ctx)
	if err != nil {
		log.Printf("[WARN] setAllConnectionStateToError failed to acquire connection from pool: %s", err.Error())
		return
	}
	defer conn.Release()

	query := connection_state.GetIncompleteConnectionStateErrorSql(connectionStateError)

	if _, err = db_local.ExecuteSqlWithArgsInTransaction(ctx, conn.Conn(), query); err != nil {
		log.Printf("[WARN] setAllConnectionStateToError failed to set connection states to error: %s", err.Error())
		return
	}
}

// OnConnectionsChanged is the callback function invoked by the connection watcher when connections are added or removed
func (state *refreshConnectionState) sendPostgreSchemaNotification(ctx context.Context) error {
	conn, err := db_local.CreateLocalDbConnection(ctx, &db_local.CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return err
	}
	notification := steampipeconfig.NewSchemaUpdateNotification(steampipeconfig.PgNotificationSchemaUpdate)

	return db_local.SendPostgresNotification(ctx, conn, notification)
}
