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
	log.Printf("[INFO] Setting up search path")
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
	utils.LogTime("db.refreshConnections start")
	defer utils.LogTime("db.refreshConnections end")

	t := time.Now()
	log.Printf("[INFO] refreshConnections start")

	// if there was an error (other than a connection error, which will NOT have been assigned to res),
	// set state of all incomplete connections to error
	defer func() {
		log.Printf("[INFO] refreshConnections complete (%fs)", time.Since(t).Seconds())
		if state.res.Error != nil {

			state.setIncompleteConnectionStateToError(ctx, sperr.WrapWithMessage(state.res.Error, "refreshConnections failed before connection update was complete"))
			// TODO send error PG notification
		}
	}()
	log.Printf("[INFO] refreshConnections building connectionUpdates")

	// determine any necessary connection updates
	state.connectionUpdates, state.res = steampipeconfig.NewConnectionUpdates(ctx, state.pool, state.forceUpdateConnectionNames...)
	defer state.logRefreshConnectionResults()
	// were we successful
	if state.res.Error != nil {
		return
	}

	log.Printf("[INFO] refreshConnections: created connection updates")

	// delete the connection state file - it will be rewritten when we are complete
	log.Printf("[INFO] refreshConnections deleting connections state file")
	steampipeconfig.DeleteConnectionStateFile()
	defer func() {
		if state.res.Error == nil {
			log.Printf("[INFO] refreshConnections saving connections state file")
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
		log.Println("[INFO] refreshConnections: no updates required")
		return
	}

	log.Printf("[INFO] refreshConnections execute connection queries")

	// execute any necessary queries
	state.executeConnectionQueries(ctx)
	if state.res.Error != nil {
		log.Printf("[WARN] refreshConnections failed with err %s", state.res.Error.Error())
		return
	}

	log.Printf("[INFO] refreshConnections complete")

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
	log.Printf("[INFO] executeConnectionQueries: num updates: %d", numUpdates)

	if numUpdates > 0 {
		// get schema queries - this updates schemas for validated plugins and drops schemas for unvalidated plugins
		state.executeUpdateQueries(ctx)
	} else if len(connectionUpdates.Delete) > 0 {
		log.Printf("[INFO] RefreshConnection has deleted all unnecessary schemas - sending notification")

		// if there are no updates and there ARE deletes, notify
		// (is there are updates, deletes will be notified by executeUpdateQueries)
		if err := state.sendPostgreSchemaNotification(ctx, maps.Keys(connectionUpdates.Delete), nil); err != nil {
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

	// retrieve updates from the table updater
	connectionUpdates := state.tableUpdater.updates

	// find any plugins which use a newer sdk version than steampipe.
	validationFailures, validatedUpdates, validatedPlugins := steampipeconfig.ValidatePlugins(connectionUpdates.Update, connectionUpdates.ConnectionPlugins)
	if len(validationFailures) > 0 {
		state.res.Warnings = append(state.res.Warnings, steampipeconfig.BuildValidationWarningString(validationFailures))
	}
	// store validated plugins in state
	// this is a map of plugins keyed by connection
	numUpdates := len(validatedUpdates)

	// we need to execute the updates in search path order
	// i.e. we first need to update the first search path connection for each plugin (this can be done in parallel)
	// then we can update the remaining connections in parallel
	// TODO make each of these an array of []ConnectionState instead of map, merge initial and dynamic
	initialUpdates, remainingUpdates, dynamicUpdates := state.populateInitialAndRemainingUpdates(validatedUpdates)

	// dynamic plugins must be updated for each plugin in search path order
	// dynamicUpdates is a map keyed by plugin with all the updates for that plugin

	// create exemplar maps
	state.exemplarSchemaMap = make(map[string]string)
	state.exemplarCommentsMap = make(map[string]string)
	log.Printf("[TRACE] executing %d update %s", numUpdates, utils.Pluralize("query", numUpdates))

	// execute initial updates
	var errors []error
	moreErrors := state.executeUpdatesAsync(ctx, initialUpdates)
	errors = append(errors, moreErrors...)

	// execute dynamic updates
	moreErrors = state.executeUpdatesAsync(ctx, dynamicUpdates)
	errors = append(errors, moreErrors...)

	// if any of the initial schemas failed, do not proceed - these schemas are required to ensure we correctly
	// resolve unqualified queries/tables
	if len(errors) > 0 {
		state.res.Error = error_helpers.CombineErrors(errors...)
		// TODO SEND ERROR NOTIFICATION
		return
	}
	log.Printf("[INFO] RefreshConnection has updated all exemplar schemas - sending notification")

	// now that we have updated all exemplar schemars, send postgres notification
	// this gives any attached interactive clients a chance to update their inspect data and autocomplete
	// (also send deletions)
	if err := state.sendPostgreSchemaNotification(ctx, maps.Keys(connectionUpdates.Delete), maps.Keys(initialUpdates)); err != nil {
		// just log
		log.Printf("[WARN] failed to send schem update Postgres notification: %s", err.Error())
	}

	// now set comments for initial updates and dynamic connections
	// note errors will be empty to get here
	moreErrors = state.UpdateCommentsAsync(ctx, initialUpdates, validatedPlugins)
	errors = append(errors, moreErrors...)

	moreErrors = state.UpdateCommentsAsync(ctx, dynamicUpdates, validatedPlugins)
	errors = append(errors, moreErrors...)

	// now execute remaining
	moreErrors = state.executeUpdatesAsync(ctx, remainingUpdates)
	errors = append(errors, moreErrors...)

	moreErrors = state.UpdateCommentsAsync(ctx, remainingUpdates, validatedPlugins)
	errors = append(errors, moreErrors...)

	if len(errors) > 0 {
		state.res.Error = error_helpers.CombineErrors(errors...)
	}

	log.Printf("[INFO] all update queries executed")

	for _, failure := range validationFailures {
		log.Printf("[TRACE] remove schema for connection failing validation connection %s, plugin Name %s\n ", failure.ConnectionName, failure.Plugin)
		if failure.ShouldDropIfExists {
			_, err := state.pool.Exec(ctx, db_common.GetDeleteConnectionQuery(failure.ConnectionName))
			if err != nil {
				// NOTE: do not return an error if we fail to remove an invalid connection - just log it
				log.Printf("[WARN] failed to delete invalid connection '%s' (%s) : %s", failure.ConnectionName, failure.Message, err.Error())
			}
		}
	}

	//if viper.GetBool(constants.ArgSchemaComments) {
	//	state.writeComments(ctx, validatedPlugins)
	//}

	log.Printf("[INFO] executeUpdateQueries complete")
	return
}

// create/update connections

func (state *refreshConnectionState) executeUpdatesAsync(ctx context.Context, updates map[string][]*steampipeconfig.ConnectionState) (errors []error) {
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

func (state *refreshConnectionState) UpdateCommentsAsync(ctx context.Context, updates map[string][]*steampipeconfig.ConnectionState, plugins map[string]*steampipeconfig.ConnectionPlugin) (errors []error) {
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

			state.updateCommentsForConnections(ctx, errChan, plugins, connectionStates)
		}(states)

	}

	wg.Wait()
	close(errChan)

	return errors
}

// syncronously execute the comments queries for one or more connections
func (state *refreshConnectionState) updateCommentsForConnections(ctx context.Context, errChan chan *connectionError, connectionPluginMap map[string]*steampipeconfig.ConnectionPlugin, connectionStates []*steampipeconfig.ConnectionState) {
	for _, connectionState := range connectionStates {
		connectionName := connectionState.ConnectionName

		var sql string

		// if this schema is static, add to the exemplar map
		state.exemplarSchemaMapMut.Lock()
		// is this plugin in the exemplarSchemaMap
		exemplarSchemaName, haveExemplarSchema := state.exemplarCommentsMap[connectionState.Plugin]
		if haveExemplarSchema {
			// we can clone!
			sql = getCloneCommentsQuery(sql, exemplarSchemaName, connectionState)
		} else {
			// get the schema from the connection plugin
			schema := connectionPluginMap[connectionName].ConnectionMap[connectionName].Schema.Schema
			// just get sql to execute update query, and update the connection state table, in a transaction
			sql = db_common.GetCommentsQueryForPlugin(connectionName, schema)
		}
		state.exemplarSchemaMapMut.Unlock()

		// the only error this will return is the failure to update the state table
		// - all other errors are written to the state table
		if err := state.executeCommentQuery(ctx, sql, connectionName); err != nil {
			errChan <- &connectionError{connectionName, err}
		} else {
			// we can clone this plugin, add to exemplarCommentsMap
			// (AFTER executing the update query)
			if !haveExemplarSchema && connectionState.CanCloneSchema() {
				state.exemplarCommentsMap[connectionState.Plugin] = connectionName
			}
		}
	}
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

func (state *refreshConnectionState) populateInitialAndRemainingUpdates(validatedUpdates steampipeconfig.ConnectionStateMap) (initialUpdates, remainingUpdates, dynamicUpdates map[string][]*steampipeconfig.ConnectionState) {
	searchPathConnections := state.connectionUpdates.FinalConnectionState.GetFirstSearchPathConnectionForPlugins(state.searchPath)
	// dynamic plugins must be updated for each plugin in search path order
	// build a map keyed by plugin, wit th evalue the ordered updates for that plugun

	// NOTE: for convenience of execution, initialUpdates and remainingUpdates are also stored as a map of []ConnectionState
	// (keyed by connection name)
	// even this there will only be one element in each array
	initialUpdates = make(map[string][]*steampipeconfig.ConnectionState)
	remainingUpdates = make(map[string][]*steampipeconfig.ConnectionState)
	dynamicUpdates = make(map[string][]*steampipeconfig.ConnectionState)

	// convert this into a lookup of initial updates to execute
	for _, connectionName := range searchPathConnections {
		if connectionState, updateRequired := validatedUpdates[connectionName]; updateRequired {
			if connectionState.SchemaMode == sdkplugin.SchemaModeDynamic {
				dynamicUpdates[connectionState.Plugin] = append(dynamicUpdates[connectionState.Plugin], connectionState)
			} else {

				initialUpdates[connectionName] = []*steampipeconfig.ConnectionState{connectionState}
			}
		}
	}
	// now add remaining updates to remainingUpdates
	for connectionName, connectionState := range validatedUpdates {
		if _, isInitialUpdate := initialUpdates[connectionName]; !isInitialUpdate {
			remainingUpdates[connectionName] = []*steampipeconfig.ConnectionState{connectionState}
		}

	}
	return initialUpdates, remainingUpdates, dynamicUpdates
}

func (state *refreshConnectionState) executeDeleteQueries(ctx context.Context) error {
	t := time.Now()
	log.Printf("[INFO] refreshConnections execute delete queries")
	defer func() {
		log.Printf("[INFO] completed execute delete queries (%fs)", time.Since(t).Seconds())
	}()

	deletions := maps.Keys(state.connectionUpdates.Delete)

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
func (state *refreshConnectionState) sendPostgreSchemaNotification(ctx context.Context, deletions, updates []string) error {
	conn, err := db_local.CreateLocalDbConnection(ctx, &db_local.CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return err
	}
	notification := steampipeconfig.NewSchemaUpdateNotification(steampipeconfig.PgNotificationSchemaUpdate)

	return db_local.SendPostgresNotification(ctx, conn, notification)
}
