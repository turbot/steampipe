package connection

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	perror_helpers "github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/introspection"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/semaphore"
)

type connectionError struct {
	name string
	err  error
}

type refreshConnectionState struct {
	// a connection pool to the DB service which uses the server appname
	pool *pgxpool.Pool
	// connectionOrder is the order of connections to be updated
	// it is the search path, with any connections NOT in the searfch path in alphabetical order at the end
	connectionOrder []string

	connectionUpdates          *steampipeconfig.ConnectionUpdates
	tableUpdater               *connectionStateTableUpdater
	res                        *steampipeconfig.RefreshConnectionResult
	forceUpdateConnectionNames []string
	// properties for schema/comment cloning
	exemplarSchemaMapMut sync.Mutex

	// maps keyed by plugin which gives an exemplar connection name,
	// if a plugin has an entry in this map, all connections schemas can be cloned from the exemplar schema
	exemplarSchemaMap map[string]string
	// if a plugin has an entry in this map, all connections schemas can be cloned from the exemplar schema
	exemplarCommentsMap map[string]string
	pluginManager       pluginManager
}

func newRefreshConnectionState(ctx context.Context, pluginManager pluginManager, forceUpdateConnectionNames []string) (*refreshConnectionState, error) {
	log.Println("[DEBUG] newRefreshConnectionState start")
	defer log.Println("[DEBUG] newRefreshConnectionState end")

	pool := pluginManager.Pool()
	if pool == nil {
		return nil, sperr.New("plugin manager returned nil pool")
	}

	// Check if GlobalConfig is initialized before proceeding
	if steampipeconfig.GlobalConfig == nil {
		return nil, sperr.New("GlobalConfig is not initialized")
	}

	// set user search path first
	log.Printf("[INFO] setting up search path")
	searchPath, err := db_local.SetUserSearchPath(ctx, pool)
	if err != nil {
		return nil, err
	}

	//build list of connections in search path order, (with non search path connections at the end)
	// get connections which are not in the search path
	nonSearchPathConnections := steampipeconfig.GlobalConfig.GetNonSearchPathConnections(searchPath)
	// sort alphabetically
	slices.Sort(nonSearchPathConnections)
	connectionOrder := append(searchPath, nonSearchPathConnections...)

	res := &refreshConnectionState{
		pool:                       pool,
		connectionOrder:            connectionOrder,
		forceUpdateConnectionNames: forceUpdateConnectionNames,
		pluginManager:              pluginManager,
	}

	return res, nil
}

// RefreshConnections loads required connections from config
// and update the database schema and search path to reflect the required connections
// return whether any changes have been made
func (s *refreshConnectionState) refreshConnections(ctx context.Context) {
	log.Println("[DEBUG] refreshConnectionState.refreshConnections start")
	defer log.Println("[DEBUG] refreshConnectionState.refreshConnections end")
	// if there was an error (other than a connection error, which will NOT have been assigned to res),
	// set state of all incomplete connections to error
	defer func() {
		if s.res != nil {
			if s.res.Error != nil {
				s.setIncompleteConnectionStateToError(ctx, sperr.WrapWithMessage(s.res.Error, "refreshConnections failed before connection update was complete"))
			}
			if !s.res.ErrorAndWarnings.Empty() {
				log.Printf("[INFO] refreshConnections completed with errors, sending notification")
				s.pluginManager.SendPostgresErrorsAndWarningsNotification(ctx, s.res.ErrorAndWarnings)
			}

		}
	}()
	log.Printf("[INFO] building connectionUpdates")

	var opts []steampipeconfig.ConnectionUpdatesOption
	if len(s.forceUpdateConnectionNames) > 0 {
		opts = append(opts, steampipeconfig.WithForceUpdate(s.forceUpdateConnectionNames))
	}

	// build a ConnectionUpdates struct
	// this determines any necessary connection updates and starts any necessary plugins
	s.connectionUpdates, s.res = steampipeconfig.NewConnectionUpdates(ctx, s.pool, s.pluginManager, opts...)

	defer s.logRefreshConnectionResults()
	// were we successful?
	if s.res.Error != nil {
		return
	}

	// if any connections in the final state are in error, that may mean we failed to start them
	// - update the connection state table
	if err := s.setFailedConnectionsToError(ctx); err != nil {
		s.res.Error = err
		return
	}

	log.Printf("[INFO] created connectionUpdates")

	//  reload plugin rate limiter definitions for all plugins which are updated - the plugin will already be loaded
	// also repopulate the plugin column table
	if err := s.updateRateLimiterDefinitions(ctx); err != nil {
		s.res.Error = err
		return
	}

	// update the plugin column table, based on connection updates and plugins with updated binaries
	if err := s.updatePluginColumnTable(ctx); err != nil {
		s.res.Error = err
		return
	}

	// delete the connection state file - it will be rewritten when we are complete
	log.Printf("[INFO] deleting connections state file")
	steampipeconfig.DeleteConnectionStateFile()
	defer func() {
		if s.res.Error == nil {
			log.Printf("[INFO] saving connections state file")
			steampipeconfig.SaveConnectionStateFile(s.res, s.connectionUpdates)
		}
	}()

	// warn about missing plugins
	s.addMissingPluginWarnings()

	// create object to update the connection state table and notify of state changes
	s.tableUpdater = newConnectionStateTableUpdater(s.connectionUpdates, s.pool)

	// NOTE: delete any DYNAMIC plugin connections which will be updated
	// to avoid them being accessed before they are updated
	log.Printf("[TRACE] deleting %d dynamic plugin connections to avoid them being accessed before they are updated", len(s.connectionUpdates.DynamicUpdates()))
	if err := s.executeDeleteQueries(ctx, s.connectionUpdates.DynamicUpdates()); err != nil {
		s.res.Error = err
		return
	}

	// update connectionState table to reflect the updates (i.e. set connections to updating/deleting/ready as appropriate)
	// also this will update the schema hashes of plugins
	if err := s.tableUpdater.start(ctx); err != nil {
		s.res.Error = err
		return
	}

	// if there are no updates, just return
	if !s.connectionUpdates.HasUpdates() {
		log.Println("[INFO] no updates required")
		return
	}

	log.Printf("[INFO] execute connection queries")

	// execute any necessary queries
	s.executeConnectionQueries(ctx)
	if s.res.Error != nil {
		log.Printf("[WARN] refreshConnections failed with err %s", s.res.Error.Error())
		return
	}

	s.res.UpdatedConnections = true
}

func (s *refreshConnectionState) setFailedConnectionsToError(ctx context.Context) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to update connection state table")
	}
	defer conn.Release()

	for _, c := range s.connectionUpdates.FinalConnectionState {
		if c.State == constants.ConnectionStateError {
			if err := s.tableUpdater.onConnectionError(ctx, conn.Conn(), c.ConnectionName, fmt.Errorf("%s", c.Error())); err != nil {
				return sperr.WrapWithMessage(err, "failed to update connection state table")
			}
		}
	}
	return nil
}

// if any plugin binaries have changed update the rate limiter definitions
func (s *refreshConnectionState) updateRateLimiterDefinitions(ctx context.Context) error {
	if len(s.connectionUpdates.PluginsWithUpdatedBinary) == 0 {
		return nil
	}

	updatedPluginLimiters, err := s.pluginManager.LoadPluginRateLimiters(s.connectionUpdates.PluginsWithUpdatedBinary)

	if err != nil {
		return err
	}

	if len(updatedPluginLimiters) > 0 {
		err := s.pluginManager.HandlePluginLimiterChanges(updatedPluginLimiters)
		if err != nil {
			s.pluginManager.SendPostgresErrorsAndWarningsNotification(ctx, perror_helpers.NewErrorsAndWarning(err))
		}
	}
	return nil
}

// if any plugin binaries have changed update the plugin column table
func (s *refreshConnectionState) updatePluginColumnTable(ctx context.Context) error {
	var deletedPlugins []string
	var updatedPlugins = map[string]*proto.Schema{}

	currentPluginConnectionMap := s.connectionUpdates.CurrentConnectionState.GetPluginToConnectionMap()
	finalPluginConnectionMap := s.connectionUpdates.FinalConnectionState.GetPluginToConnectionMap()

	// add into plugin column table any plugins which have connections for the first time
	for _, connectionState := range s.connectionUpdates.Update {
		connectionName := connectionState.ConnectionName
		if connectionState.SchemaMode == plugin.SchemaModeDynamic {
			// plugin column table only supports static for now
			continue
		}
		p := connectionState.Plugin
		if _, ok := currentPluginConnectionMap[p]; !ok {
			updatedPlugins[p] = s.connectionUpdates.ConnectionPlugins[connectionName].ConnectionMap[connectionName].Schema
		}
	}

	// remove from plugin column table any plugins which have no connections
	for connectionName := range s.connectionUpdates.Delete {
		// get plugin for this connection
		connectionState, ok := s.connectionUpdates.CurrentConnectionState[connectionName]
		if !ok {
			continue
		}

		p := connectionState.Plugin
		if _, ok := finalPluginConnectionMap[p]; !ok {
			deletedPlugins = append(deletedPlugins, p)
		}
	}

	// update plugin column table for any plugins which have updated binaries
	for p, connectionName := range s.connectionUpdates.PluginsWithUpdatedBinary {
		// do we actually have a connection plugin for this plugin?
		if connectionPlugin, ok := s.connectionUpdates.ConnectionPlugins[connectionName]; ok {
			updatedPlugins[p] = connectionPlugin.ConnectionMap[connectionName].Schema
		}
	}

	return s.pluginManager.UpdatePluginColumnsTable(ctx, updatedPlugins, deletedPlugins)

}

func (s *refreshConnectionState) addMissingPluginWarnings() {
	log.Printf("[INFO] refreshConnections: identify missing plugins")

	var connectionNames []string
	// add warning if there are connections left over, from missing plugins
	if len(s.connectionUpdates.MissingPlugins) > 0 {
		// warning
		for _, conns := range s.connectionUpdates.MissingPlugins {
			for _, con := range conns {
				connectionNames = append(connectionNames, con.Name)
			}

		}
		pluginNames := maps.Keys(s.connectionUpdates.MissingPlugins)

		s.res.AddWarning(fmt.Sprintf("%d %s required by %d %s %s missing. To install, please run: %s",
			len(pluginNames),
			utils.Pluralize("plugin", len(pluginNames)),
			len(connectionNames),
			utils.Pluralize("connection", len(connectionNames)),
			utils.Pluralize("is", len(pluginNames)),
			pconstants.Bold(fmt.Sprintf("steampipe plugin install %s", strings.Join(pluginNames, " ")))))
	}
}

func (s *refreshConnectionState) logRefreshConnectionResults() {
	// Safe type assertion to avoid panic if viper.Get returns nil or wrong type
	cmdValue := viper.Get(constants.ConfigKeyActiveCommand)
	if cmdValue == nil {
		return
	}

	cmd, ok := cmdValue.(*cobra.Command)
	if !ok || cmd == nil {
		return
	}

	cmdName := cmd.Name()
	if cmdName != "plugin-manager" {
		return
	}

	var op strings.Builder
	if s.connectionUpdates != nil {
		op.WriteString(s.connectionUpdates.String())
	}
	if s.res != nil {
		op.WriteString(fmt.Sprintf("%s\n", s.res.String()))
	}

	log.Printf("[TRACE] refresh connections: \n%s\n", helpers.Tabify(op.String(), "    "))
}

func (s *refreshConnectionState) executeConnectionQueries(ctx context.Context) {
	log.Println("[DEBUG] refreshConnectionState.executeConnectionQueries start")
	defer log.Println("[DEBUG] refreshConnectionState.executeConnectionQueries end")

	// execute deletions
	if err := s.executeDeleteQueries(ctx, s.connectionUpdates.GetConnectionsToDelete()); err != nil {
		// just log
		log.Printf("[WARN] failed to delete all unused schemas: %s", err.Error())
	}

	// execute updates
	numUpdates := len(s.connectionUpdates.Update)
	numMissingComments := len(s.connectionUpdates.MissingComments)
	log.Printf("[INFO] executeConnectionQueries: num updates: %d, connections missing comments: %d", numUpdates, numMissingComments)

	if numUpdates+numMissingComments > 0 {
		// get schema queries - this updates schemas for validated plugins and drops schemas for unvalidated plugins
		s.executeUpdateQueries(ctx)
		// done
		return
	}

	if len(s.connectionUpdates.Delete) > 0 {
		log.Printf("[INFO] deleted all unnecessary schemas - sending notification")

		// if there are no updates and there ARE deletes, notify
		// (is there are updates, deletes will be notified by executeUpdateQueries)
		if err := s.pluginManager.SendPostgresSchemaNotification(ctx); err != nil {
			// just log
			log.Printf("[WARN] failed to send schema deletion Postgres notification: %s", err.Error())
		}
	}
}

// execute all update queries
// NOTE: this only sets res.Error if there is a failure to set update the connection state table
// - all other connection based failures are recorded in the connection state table
func (s *refreshConnectionState) executeUpdateQueries(ctx context.Context) {
	log.Println("[DEBUG] refreshConnectionState.executeUpdateQueries start")
	defer log.Println("[DEBUG] refreshConnectionState.executeUpdateQueries end")

	defer func() {
		if s.res.Error != nil {
			log.Printf("[INFO] executeUpdateQueries returned error: %v", s.res.Error)
		}
	}()

	connectionUpdates := s.connectionUpdates
	connectionPlugins := connectionUpdates.ConnectionPlugins
	numUpdates := len(connectionUpdates.Update)

	// we need to execute the updates in search path order
	// i.e. we first need to update the first search path connection for each plugin (this can be done in parallel)
	// then we can update the remaining connections in parallel
	initialUpdates, remainingUpdates, dynamicUpdates := s.getInitialAndRemainingUpdates()

	// dynamic plugins must be updated for each plugin in search path order
	// dynamicUpdates is a map keyed by plugin with all the updates for that plugin

	// create exemplar maps
	s.exemplarSchemaMap = make(map[string]string)
	s.exemplarCommentsMap = make(map[string]string)
	log.Printf("[INFO] executing %d update %s", numUpdates, utils.Pluralize("query", numUpdates))

	// execute initial updates
	var errors []error
	if len(initialUpdates) > 0 {
		log.Printf("[INFO] executing %d initial %s", len(initialUpdates), utils.Pluralize("update", len(initialUpdates)))
		moreErrors := s.executeUpdatesInParallel(ctx, initialUpdates)
		errors = append(errors, moreErrors...)
	}

	if len(dynamicUpdates) > 0 {
		// execute dynamic updates (note, we update all connections in search path order,
		// so must call executeUpdateSetsInParallel)
		log.Printf("[INFO] executing %d dynamic %s", len(dynamicUpdates), utils.Pluralize("update", len(dynamicUpdates)))
		moreErrors := s.executeUpdateSetsInParallel(ctx, dynamicUpdates)
		errors = append(errors, moreErrors...)
	}

	// if any of the initial schemas failed, do not proceed - these schemas are required to ensure we correctly
	// resolve unqualified queries/tables
	if len(errors) > 0 {
		s.res.Error = error_helpers.CombineErrors(errors...)
		log.Printf("[WARN] initial updates failed: %s", s.res.Error.Error())
		return
	}

	log.Printf("[INFO] set comments for initial updates")
	// now set comments for initial updates and dynamic connections
	// note errors will be empty to get here
	s.UpdateCommentsInParallel(ctx, maps.Values(initialUpdates), connectionPlugins)

	log.Printf("[INFO] set comments for dynamic updates")
	// convert dynamicUpdates to an array of connection states
	var dynamicUpdateArray = updateSetMapToArray(dynamicUpdates)
	s.UpdateCommentsInParallel(ctx, dynamicUpdateArray, connectionPlugins)

	log.Printf("[INFO] updated all exemplar schemas - sending notification")
	// now that we have updated all exemplar schemars, send postgres notification
	// this gives any attached interactive clients a chance to update their inspect data and autocomplete
	if err := s.pluginManager.SendPostgresSchemaNotification(ctx); err != nil {
		// just log
		log.Printf("[WARN] failed to send schem update Postgres notification: %s", err.Error())
	}

	if len(remainingUpdates) > 0 {
		log.Printf("[INFO] Execute %d remaining %s",
			len(remainingUpdates),
			utils.Pluralize("updates", len(remainingUpdates)))
		// now execute remaining updates
		moreErrors := s.executeUpdatesInParallel(ctx, remainingUpdates)
		errors = append(errors, moreErrors...)
	}

	log.Printf("[INFO] Set comments for %d remaining %s and %d %s missing comments",
		len(remainingUpdates),
		utils.Pluralize("updates", len(remainingUpdates)),
		len(connectionUpdates.MissingComments),
		utils.Pluralize("updates", len(connectionUpdates.MissingComments)),
	)
	// set comments for remaining updates
	s.UpdateCommentsInParallel(ctx, maps.Values(remainingUpdates), connectionPlugins)
	// set comments for any other connection without comment set
	s.UpdateCommentsInParallel(ctx, maps.Values(s.connectionUpdates.MissingComments), connectionPlugins)

	if len(errors) > 0 {
		s.res.Error = error_helpers.CombineErrors(errors...)
	}

	log.Printf("[INFO] all update queries executed")

	for _, failure := range connectionUpdates.InvalidConnections {
		log.Printf("[TRACE] remove schema for connection failing validation connection %s, plugin Name %s\n ", failure.ConnectionName, failure.Plugin)
		if failure.ShouldDropIfExists {
			_, err := s.pool.Exec(ctx, db_common.GetDeleteConnectionQuery(failure.ConnectionName))
			if err != nil {
				// NOTE: do not return an error if we fail to remove an invalid connection - just log it
				log.Printf("[WARN] failed to delete invalid connection '%s' (%s) : %s", failure.ConnectionName, failure.Message, err.Error())
			}
		}
	}
	log.Printf("[INFO] executeUpdateQueries complete")
	return
}

// convert map update sets (used for dynamic schemas) to an array of the underlying connection states
func updateSetMapToArray(updateSetMap map[string][]*steampipeconfig.ConnectionState) []*steampipeconfig.ConnectionState {
	var res []*steampipeconfig.ConnectionState
	for _, updates := range updateSetMap {
		res = append(res, updates...)
	}
	return res
}

// create/update connections

func (s *refreshConnectionState) executeUpdatesInParallel(ctx context.Context, updates map[string]*steampipeconfig.ConnectionState) (errors []error) {
	log.Println("[DEBUG] refreshConnectionState.executeUpdatesInParallel start")
	defer log.Println("[DEBUG] refreshConnectionState.executeUpdatesInParallel end")

	// convert updates to update sets
	updatesAsSets := make(map[string][]*steampipeconfig.ConnectionState, len(updates))
	for k, v := range updates {
		updatesAsSets[k] = []*steampipeconfig.ConnectionState{v}
	}
	// just call executeUpdateSetsInParallel
	return s.executeUpdateSetsInParallel(ctx, updatesAsSets)
}

// execute sets of updates in parallel - this is required as for dynamic plugins, we must update all connections in
// search path order
// - for convenience we also use this function for static connections by mapping the input data
// from map[string]*steampipeconfig.ConnectionState to map[string][]*steampipeconfig.ConnectionState
func (s *refreshConnectionState) executeUpdateSetsInParallel(ctx context.Context, updates map[string][]*steampipeconfig.ConnectionState) (errors []error) {
	log.Println("[DEBUG] refreshConnectionState.executeUpdateSetsInParallel start")
	defer log.Println("[DEBUG] refreshConnectionState.executeUpdateSetsInParallel end")

	var wg sync.WaitGroup
	var errChan = make(chan *connectionError)

	// default to running a single update at a time
	var maxParallel = int64(1)
	// allow override of this behaviour vis env var
	if envMaxStr, ok := os.LookupEnv("STEAMPIPE_UPDATE_SCHEMA_MAX_PARALLEL"); ok {
		envMax, err := strconv.Atoi(envMaxStr)
		if err == nil {
			maxParallel = int64(envMax)
		}
	}
	log.Printf("[INFO] executeUpdateSetsInParallel - maxParallel= %d", maxParallel)

	sem := semaphore.NewWeighted(maxParallel)

	go func() {
		for connectionError := range errChan {
			errors = append(errors, connectionError.err)
			conn, poolErr := s.pool.Acquire(ctx)
			if poolErr == nil {
				if err := s.tableUpdater.onConnectionError(ctx, conn.Conn(), connectionError.name, connectionError.err); err != nil {
					log.Println("[WARN] failed to update connection state table", err.Error())
				}
				conn.Release()
			}
		}
	}()

	// allow disabling of schema clone via env var
	var cloneSchemaEnabled = true
	if envClone, ok := os.LookupEnv("STEAMPIPE_CLONE_SCHEMA"); ok {
		cloneSchemaEnabled = strings.ToLower(envClone) == "true"
	}
	log.Printf("[INFO] executeUpdateForConnections - cloneSchemaEnabled=%v", cloneSchemaEnabled)

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

			// Check if context is cancelled before starting work
			select {
			case <-ctx.Done():
				// Context cancelled - don't process this batch
				return
			default:
				// Context still valid - proceed with work
			}

			s.executeUpdateForConnections(ctx, errChan, cloneSchemaEnabled, connectionStates...)
		}(states)

	}

	wg.Wait()
	close(errChan)

	return errors
}

// syncronously execute the update queries for one or more connections
func (s *refreshConnectionState) executeUpdateForConnections(ctx context.Context, errChan chan *connectionError, cloneSchemaEnabled bool, connectionStates ...*steampipeconfig.ConnectionState) {
	log.Println("[DEBUG] refreshConnectionState.executeUpdateForConnections start")
	defer log.Println("[DEBUG] refreshConnectionState.executeUpdateForConnections end")

	for _, connectionState := range connectionStates {
		// Check if context is cancelled before processing each connection
		select {
		case <-ctx.Done():
			// Context cancelled - stop processing remaining connections
			log.Println("[DEBUG] context cancelled, stopping executeUpdateForConnections")
			return
		default:
			// Context still valid - continue
		}

		connectionName := connectionState.ConnectionName
		pluginSchemaName := utils.PluginFQNToSchemaName(connectionState.Plugin)
		var sql string

		s.exemplarSchemaMapMut.Lock()
		// is this plugin in the exemplarSchemaMap
		exemplarSchemaName, haveExemplarSchema := s.exemplarSchemaMap[connectionState.Plugin]
		if haveExemplarSchema && cloneSchemaEnabled {
			// we can clone!
			sql = getCloneSchemaQuery(exemplarSchemaName, connectionState)
		} else {
			// just get sql to execute update query, and update the connection state table, in a transaction
			sql = db_common.GetUpdateConnectionQuery(connectionName, pluginSchemaName)
		}
		s.exemplarSchemaMapMut.Unlock()

		// the only error this will return is the failure to update the state table
		// - all other errors are written to the state table
		if err := s.executeUpdateQuery(ctx, sql, connectionName); err != nil {
			errChan <- &connectionError{connectionName, err}
		} else {
			// we can clone this plugin, add to exemplarSchemaMap
			// (AFTER executing the update query)
			if !haveExemplarSchema && connectionState.CanCloneSchema() {
				// Fix #4757: Protect map write with mutex to prevent race condition
				s.exemplarSchemaMapMut.Lock()
				s.exemplarSchemaMap[connectionState.Plugin] = connectionName
				s.exemplarSchemaMapMut.Unlock()
			}
		}
	}
}

func (s *refreshConnectionState) executeUpdateQuery(ctx context.Context, sql, connectionName string) (err error) {
	log.Println("[DEBUG] refreshConnectionState.executeUpdateQuery start")
	defer log.Println("[DEBUG] refreshConnectionState.executeUpdateQuery end")

	// create a transaction
	tx, err := s.pool.Begin(ctx)
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
		s.res.AddFailedConnection(connectionName, err.Error())

		// update the state table
		//(the transaction will be aborted - create a connection for the update)
		if conn, poolErr := s.pool.Acquire(ctx); poolErr == nil {
			defer conn.Release()
			if statusErr := s.tableUpdater.onConnectionError(ctx, conn.Conn(), connectionName, err); statusErr != nil {
				// NOTE: do not return the error - unless we failed to update the connection state table
				return error_helpers.CombineErrorsWithPrefix(fmt.Sprintf("failed to update connection %s and failed to update connection_state table", connectionName), err, statusErr)
			}
		}
		return nil
	}

	// update state table (inside transaction)
	err = s.tableUpdater.onConnectionReady(ctx, tx.Conn(), connectionName)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to update connection state table")
	}
	return nil
}

// set connection comments

func (s *refreshConnectionState) UpdateCommentsInParallel(ctx context.Context, updates []*steampipeconfig.ConnectionState, plugins map[string]*steampipeconfig.ConnectionPlugin) (errors []error) {
	if !viper.GetBool(pconstants.ArgSchemaComments) {
		return nil
	}

	var wg sync.WaitGroup
	var errChan = make(chan *connectionError)

	// use as many goroutines as we have connections
	var maxUpdateThreads = int64(s.pool.Config().MaxConns)
	sem := semaphore.NewWeighted(maxUpdateThreads)

	go func() {
		for {
			select {
			case connectionError := <-errChan:
				if connectionError == nil {
					return
				}
				errors = append(errors, connectionError.err)
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

			s.updateCommentsForConnection(ctx, errChan, plugins, connectionState)
		}(connectionState)

	}

	wg.Wait()
	close(errChan)

	return errors
}

// syncronously execute the comments queries for one or more connections
func (s *refreshConnectionState) updateCommentsForConnection(ctx context.Context, errChan chan *connectionError, connectionPluginMap map[string]*steampipeconfig.ConnectionPlugin, connectionState *steampipeconfig.ConnectionState) {
	log.Printf("[DEBUG] refreshConnectionState.updateCommentsForConnection start for connection '%s'", connectionState.ConnectionName)

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
	if err := s.executeCommentQuery(ctx, sql, connectionName); err != nil {
		errChan <- &connectionError{connectionName, err}
	} //else {
	//	// we can clone this plugin, add to exemplarCommentsMap
	//	// (AFTER executing the update query)
	//	if !haveExemplarSchema && connectionState.CanCloneSchema() {
	//		state.exemplarCommentsMap[connectionState.Plugin] = connectionName
	//	}
	//}
}

func (s *refreshConnectionState) executeCommentQuery(ctx context.Context, sql, connectionName string) error {
	// create a transaction
	tx, err := s.pool.Begin(ctx)
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
		if conn, poolErr := s.pool.Acquire(ctx); poolErr == nil {
			defer conn.Release()
			if statusErr := s.tableUpdater.onConnectionError(ctx, conn.Conn(), connectionName, err); statusErr != nil {
				// NOTE: do not return the error - unless we failed to update the connection state table
				return error_helpers.CombineErrorsWithPrefix(fmt.Sprintf("failed to update connection %s and failed to update connection_state table", connectionName), err, statusErr)
			}
		}

		return nil
	}

	// update state table (inside transaction)
	// ignore error
	if err := s.tableUpdater.onConnectionCommentsLoaded(ctx, tx.Conn(), connectionName); err != nil {
		log.Printf("[WARN] failed to set 'comments_set' for connection '%s': %s", connectionName, err.Error())
	}

	return nil
}

func getCloneSchemaQuery(exemplarSchemaName string, connectionState *steampipeconfig.ConnectionState) string {
	return fmt.Sprintf("select clone_foreign_schema('%s', '%s', '%s');", exemplarSchemaName, connectionState.ConnectionName, connectionState.Plugin)
}

func (s *refreshConnectionState) getInitialAndRemainingUpdates() (initialUpdates, remainingUpdates map[string]*steampipeconfig.ConnectionState, dynamicUpdates map[string][]*steampipeconfig.ConnectionState) {
	updates := s.connectionUpdates.Update
	searchPathConnections := s.connectionUpdates.FinalConnectionState.GetFirstSearchPathConnectionForPlugins(s.connectionOrder)

	initialUpdates = make(map[string]*steampipeconfig.ConnectionState)
	remainingUpdates = make(map[string]*steampipeconfig.ConnectionState)
	// dynamic plugins must be updated for each plugin in search path order
	// build a map keyed by plugin, with the value the ordered updates for that plugin
	dynamicUpdates = make(map[string][]*steampipeconfig.ConnectionState)

	// convert this into a lookup of initial updates to execute
	for _, connectionName := range searchPathConnections {
		if connectionState, updateRequired := updates[connectionName]; updateRequired {
			if connectionState.SchemaMode == plugin.SchemaModeDynamic {
				pluginInstance := *connectionState.PluginInstance
				dynamicUpdates[pluginInstance] = append(dynamicUpdates[pluginInstance], connectionState)
			} else {
				initialUpdates[connectionName] = connectionState
			}
		}
	}
	// now add remaining updates to remainingUpdates
	for connectionName, connectionState := range updates {
		_, isInitialUpdate := initialUpdates[connectionName]
		if connectionState.SchemaMode == plugin.SchemaModeStatic && !isInitialUpdate {
			remainingUpdates[connectionName] = connectionState
		}
	}

	log.Printf("[TRACE] getInitialAndRemainingUpdates: %d initialUpdates: %s, %d remainingUpdates: %s, %d dynamicUpdates: %s",
		len(initialUpdates),
		strings.Join(maps.Keys(initialUpdates), ", "),
		len(remainingUpdates),
		strings.Join(maps.Keys(remainingUpdates), ", "),
		len(dynamicUpdates),
		strings.Join(maps.Keys(dynamicUpdates), ", "))

	if len(initialUpdates)+len(dynamicUpdates)+len(remainingUpdates) != len(updates) {
		log.Printf("[WARN] getInitialAndRemainingUpdates: initialUpdates + remainingUpdates + dynamicUpdates != updates")
	}

	return initialUpdates, remainingUpdates, dynamicUpdates
}

func (s *refreshConnectionState) executeDeleteQueries(ctx context.Context, deletions []string) error {
	t := time.Now()
	log.Printf("[INFO] execute %d delete %s", len(deletions), utils.Pluralize("query", len(deletions)))
	defer func() {
		log.Printf("[INFO] completed execute delete queries (%fs)", time.Since(t).Seconds())
	}()

	var errors []error

	for _, c := range deletions {
		err := s.executeDeleteQuery(ctx, c)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return error_helpers.CombineErrors(errors...)
}

// delete the schema and update remove the connection from the state table
// NOTE: this only returns an error if we fail to update the state table
func (s *refreshConnectionState) executeDeleteQuery(ctx context.Context, connectionName string) error {
	// create a transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to create transaction to perform delete query")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	sql := db_common.GetDeleteConnectionQuery(connectionName)

	// execute delete sql
	_, err = tx.Exec(ctx, sql)
	if err != nil {
		// update the state table
		//(the transaction will be aborted - create a connection for the update)
		if conn, poolErr := s.pool.Acquire(ctx); poolErr == nil {
			defer conn.Release()
			if statusErr := s.tableUpdater.onConnectionError(ctx, conn.Conn(), connectionName, err); statusErr != nil {
				// NOTE: do not return the error - unless we failed to update the connection state table
				return error_helpers.CombineErrorsWithPrefix(fmt.Sprintf("failed to update connection %s and failed to update connection_state table", connectionName), err, statusErr)
			}
		}

		return nil
	}

	// delete state table entry (inside transaction)
	err = s.tableUpdater.onConnectionDeleted(ctx, tx.Conn(), connectionName)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to delete connection state table entry for '%s'", connectionName)
	}
	return nil
}

// set the state of any incomplete connections to error
func (s *refreshConnectionState) setIncompleteConnectionStateToError(ctx context.Context, err error) {
	// create wrapped error
	connectionStateError := sperr.WrapWithMessage(err, "failed to update Steampipe connections")
	// load connection state
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		log.Printf("[WARN] setAllConnectionStateToError failed to acquire connection from pool: %s", err.Error())
		return
	}
	defer conn.Release()

	queries := introspection.GetIncompleteConnectionStateErrorSql(connectionStateError)

	if _, err = db_local.ExecuteSqlWithArgsInTransaction(ctx, conn.Conn(), queries...); err != nil {
		log.Printf("[WARN] setAllConnectionStateToError failed to set connection states to error: %s", err.Error())
		return
	}
}
