package db_local

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
	"golang.org/x/sync/semaphore"
	"log"
	"strings"
	"sync"
)

func RefreshConnectionAndSearchPaths(ctx context.Context, forceUpdateConnectionNames ...string) *steampipeconfig.RefreshConnectionResult {
	// uncomment to debug
	//time.Sleep(10 * time.Second)

	// create a connection pool to connection refresh
	poolsize := 1
	pool, err := createConnectionPool(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser}, poolsize)
	if err != nil {
		return steampipeconfig.NewErrorRefreshConnectionResult(err)
	}
	defer pool.Close()

	// set user search path first
	log.Printf("[TRACE] Setting up search path")
	searchPath, err := setUserSearchPath(ctx, pool)
	if err != nil {
		return steampipeconfig.NewErrorRefreshConnectionResult(err)
	}

	// now refresh connections
	log.Printf("[TRACE] Refreshing connections")
	res := refreshConnections(ctx, pool, searchPath, forceUpdateConnectionNames...)
	if res.Error != nil {
		return res
	}

	return res
}

// RefreshConnections loads required connections from config
// and update the database schema and search path to reflect the required connections
// return whether any changes have been made
func refreshConnections(ctx context.Context, pool *pgxpool.Pool, searchPath []string, forceUpdateConnectionNames ...string) (res *steampipeconfig.RefreshConnectionResult) {
	log.Printf("[INFO] refreshConnections")
	//
	utils.LogTime("db.refreshConnections start")
	defer utils.LogTime("db.refreshConnections end")

	defer func() {
		if res.Error != nil {
			// if there was an error, set state of all connectoins to error
			// TODO KAI CHECK THIS
			setAllConnectionStateToError(ctx, pool, res.Error)
			// TODO kai send error PG notification
		}
	}()
	// determine any necessary connection updates
	connectionUpdates, res := steampipeconfig.NewConnectionUpdates(ctx, pool, forceUpdateConnectionNames...)
	defer logRefreshConnectionResults(connectionUpdates, res)
	if res.Error != nil {
		return res
	}

	log.Printf("[INFO] refreshConnections: created connection updates")

	var connectionNames, pluginNames []string
	// add warning if there are connections left over, from missing plugins
	if len(connectionUpdates.MissingPlugins) > 0 {
		// warning
		for a, conns := range connectionUpdates.MissingPlugins {
			for _, con := range conns {
				connectionNames = append(connectionNames, con.Name)
			}
			pluginNames = append(pluginNames, utils.GetPluginName(a))
		}
		res.AddWarning(fmt.Sprintf("%d %s required by %s %s missing. To install, please run %s",
			len(pluginNames),
			utils.Pluralize("plugin", len(pluginNames)),
			utils.Pluralize("connection", len(connectionNames)),
			utils.Pluralize("is", len(pluginNames)),
			constants.Bold(fmt.Sprintf("steampipe plugin install %s", strings.Join(pluginNames, " ")))))
	}

	// create object to update the connection state table and notify of state changes
	tableUpdater := newConnectionStateTableUpdater(connectionUpdates)

	// update connectionState table to reflect the updates (i.e. set connections to updating/deleting/ready as appropriate)
	// also this will update the schema hashes of plugins
	if err := tableUpdater.start(ctx); err != nil {
		res.Error = err
		return res
	}

	// if there ar eno updates, just return
	if !connectionUpdates.HasUpdates() {
		log.Println("[INFO] refreshConnections: no updates required")
		return res
	}

	// delete the connection state file - this indicates to anything using it that we are in the process up refreshing
	log.Printf("[TRACE] refreshConnections deleting connections state file")
	steampipeconfig.DeleteConnectionStateFile()

	// before finishing - be sure to save connection state if there was no error
	defer func() {
		if res.Error == nil {
			log.Printf("[INFO] refreshConnections saving connections state file")
			// now serialise the connection state
			if res.Error == nil {
				steampipeconfig.SaveConnectionStateFile(res, connectionUpdates)
			}
		}
	}()

	log.Printf("[INFO] refreshConnections execute connection queries")

	// now build list of necessary queries to perform the update
	queryRes := executeConnectionQueries(ctx, pool, searchPath, tableUpdater)
	// merge results into local results
	res.Merge(queryRes)
	if res.Error != nil {
		return res
	}

	res.UpdatedConnections = true

	return res
}

// sett the state of all connections to error
func setAllConnectionStateToError(ctx context.Context, pool *pgxpool.Pool, err error) {
	// create wrapped error
	connectionStateError := sperr.WrapWithMessage(err, "failed to update Steampipe connections")
	// load connection state
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("[WARN] setAllConnectionStateToError failed to acquire conneciton from pool: %s", err.Error())
	}
	defer conn.Release()

	// load the connection state file and filter out any connections which are not in the list of schemas
	// this allows for the database being rebuilt,modified externally
	currentConnectionState, err := steampipeconfig.LoadConnectionState(ctx, conn.Conn())
	if err != nil {
		log.Printf("[WARN] setAllConnectionStateToError failed to load connection state: %s", err.Error())
		return
	}
	var queries []db_common.QueryWithArgs
	for name := range currentConnectionState {
		queries = append(queries, getConnectionStateErrorSql(name, connectionStateError))
	}
	if _, err := executeSqlWithArgsAsRoot(ctx, queries...); err != nil {
		log.Printf("[WARN] setAllConnectionStateToError failed to set connectoin state to error: %s", err.Error())
	}
}

func logRefreshConnectionResults(updates *steampipeconfig.ConnectionUpdates, res *steampipeconfig.RefreshConnectionResult) {
	var cmdName = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command).Name()
	if cmdName != "plugin-manager" {
		return
	}

	var op strings.Builder
	if updates != nil {
		op.WriteString(fmt.Sprintf("%s", updates.String()))
	}
	if res != nil {
		op.WriteString(fmt.Sprintf("%s\n", res.String()))
	}

	log.Printf("[INFO] refresh connections: \n%s\n", helpers.Tabify(op.String(), "    "))
}

func executeConnectionQueries(ctx context.Context, pool *pgxpool.Pool, searchPath []string, tableUpdater *connectionStateTableUpdater) *steampipeconfig.RefreshConnectionResult {
	// retrieve updates from the table updater
	connectionUpdates := tableUpdater.updates

	utils.LogTime("db.executeConnectionQueries start")
	defer utils.LogTime("db.executeConnectionQueries start")

	numUpdates := len(connectionUpdates.Update)
	log.Printf("[TRACE] executeConnectionQueries: num updates %d", numUpdates)

	res := &steampipeconfig.RefreshConnectionResult{}
	if numUpdates > 0 {
		// get schema queries - this updates schemas for validated plugins and drops schemas for unvalidated plugins
		res = executeUpdateQueries(ctx, pool, searchPath, tableUpdater)
		if res.Error != nil {
			log.Printf("[TRACE] executeUpdateQueries returned error: %v", res.Error)
			return res
		}
	}

	// delete connections
	executeDeleteQueries(ctx, pool, connectionUpdates.Delete, tableUpdater)

	return res
}

func executeUpdateQueries(ctx context.Context, pool *pgxpool.Pool, searchPath []string, tableUpdater *connectionStateTableUpdater) (res *steampipeconfig.RefreshConnectionResult) {
	utils.LogTime("db.executeUpdateQueries start")
	defer utils.LogTime("db.executeUpdateQueries end")

	res = &steampipeconfig.RefreshConnectionResult{}

	// retrieve updates from the table updater
	connectionUpdates := tableUpdater.updates

	// find any plugins which use a newer sdk version than steampipe.
	validationFailures, validatedUpdates, validatedPlugins := steampipeconfig.ValidatePlugins(connectionUpdates.Update, connectionUpdates.ConnectionPlugins)
	if len(validationFailures) > 0 {
		res.Warnings = append(res.Warnings, steampipeconfig.BuildValidationWarningString(validationFailures))
	}
	numUpdates := len(validatedUpdates)

	//we need to execute the updates in search path order

	// group the search path by plugin, then add/update the first connection for each plugin together, first
	pluginSearchPathMap := make(map[string][]string)
	var initialUpdates []string
	var remainingUpdates []string

	//orderedUpdates := make([]string, 0, numUpdates)

	// build ordered updated
	for _, connectionName := range searchPath {
		// get the connection config for this
		connectionConfig, gotConfig := steampipeconfig.GlobalConfig.Connections[connectionName]
		// do we have config for this search path element (if not just ignore)
		if gotConfig {
			// add to pluginSearchPathMap for this plugin
			pluginSearchPathMap[connectionConfig.Plugin] = append(pluginSearchPathMap[connectionConfig.Plugin], connectionName)
		}
	}

	// now construct ordered updates
	// build a list of upates which are the first conneciton for each plugin
	// (these can be executed in parallel, but must be executed first)
	// and a list of all other updates
	for _, connections := range pluginSearchPathMap {
		for i, connectionName := range connections {
			// is an update required for this connection
			if _, updateRequired := validatedUpdates[connectionName]; updateRequired {
				// if an update is required for first plugin, add to initialUpdates
				if i == 0 {
					initialUpdates = append(initialUpdates, connectionName)
				} else {
					remainingUpdates = append(remainingUpdates, connectionName)
				}
			}
		}
	}

	// add search path first
	//if _, updateConnection := validatedUpdates[c]; updateConnection {
	//	orderedUpdates = append(orderedUpdates, c)
	//}

	// now add all updates NOT in the search path (if any)
	if len(initialUpdates)+len(remainingUpdates) < numUpdates {
		// build map from search path
		searchPathMap := utils.SliceToLookup(searchPath)

		for c := range validatedUpdates {
			if _, inSearchPath := searchPathMap[c]; !inSearchPath {
				remainingUpdates = append(remainingUpdates, c)
			}
		}
	}

	exemplarSchemaMap := make(map[string]string)
	log.Printf("[TRACE] executing %d update %s", numUpdates, utils.Pluralize("query", numUpdates))

	// TODO kai paralellizeq
	var errors []error
	for _, connectionName := range initialUpdates {
		connectionData := validatedUpdates[connectionName]
		remoteSchema := utils.PluginFQNToSchemaName(connectionData.Plugin)
		// TODO KAI NOTE for now we ignore cloning
		// if this schema is static, add to the exemplar map
		connectionData.CanCloneSchema()
		{
			exemplarSchemaMap[connectionData.Plugin] = connectionName
		}

		// execute update query, and update the connection state table, in a transaction
		sql := getUpdateConnectionQuery(connectionName, remoteSchema)
		if err := executeUpdateQuery(ctx, pool, tableUpdater, sql, connectionName); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		res.Error = error_helpers.CombineErrors(errors...)
		return res
	}

	// now execute remaining
	// TODO KAI wrap this in parallel function which either clones or not
	for _, connectionName := range remainingUpdates {
		connectionData := validatedUpdates[connectionName]
		remoteSchema := utils.PluginFQNToSchemaName(connectionData.Plugin)
		// execute update query, and update the connection state table, in a transaction
		sql := getUpdateConnectionQuery(connectionName, remoteSchema)
		if err := executeUpdateQuery(ctx, pool, tableUpdater, sql, connectionName); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		res.Error = error_helpers.CombineErrors(errors...)
	}
	//	statushooks.SetStatus(ctx, fmt.Sprintf("Cloning %d %s", len(cloneableConnections), utils.Pluralize("connection", len(cloneableConnections))))
	//	if err := cloneConnectionSchemas(ctx, pool, exemplarSchemaMap, cloneableConnections, idx, numUpdates, tableUpdater); err != nil {
	//		res.Error = err
	//		return res
	//	}
	//}

	log.Printf("[TRACE] all update queries executed")

	for _, failure := range validationFailures {
		log.Printf("[TRACE] remove schema for connection failing validation connection %s, plugin Name %s\n ", failure.ConnectionName, failure.Plugin)
		if failure.ShouldDropIfExists {
			_, err := pool.Exec(ctx, getDeleteConnectionQuery(failure.ConnectionName))
			if err != nil {
				errors = append(errors, err)
			}
		}
	}
	if len(errors) > 0 {
		res.Error = error_helpers.CombineErrors(errors...)
		return res
	}

	if viper.GetBool(constants.ArgSchemaComments) {
		log.Printf("[WARN] start comments")

		conn, err := pool.Acquire(ctx)
		if err != nil {
			log.Printf("[WARN] comments error %v", err)
			// todo send error notification
			res.Error = err
			return res
		}
		defer conn.Release()
		numCommentsUpdates := len(validatedPlugins)
		log.Printf("[TRACE] executing %d comment %s", numCommentsUpdates, utils.Pluralize("query", numCommentsUpdates))

		for connectionName, connectionPlugin := range validatedPlugins {
			_, err = executeSqlInTransaction(ctx, conn.Conn(), "lock table pg_namespace;", getCommentsQueryForPlugin(connectionName, connectionPlugin))
			if err != nil {
				// todo send error notification
				res.Error = err
				return res
			}
		}
	}

	log.Printf("[TRACE] executeUpdateQueries complete")
	return res
}

func executeUpdateQuery(ctx context.Context, pool *pgxpool.Pool, tableUpdater *connectionStateTableUpdater, sql, connectionName string) (err error) {
	// create a transaction
	tx, err := pool.Begin(ctx)
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
		tableUpdater.onConnectionError(ctx, tx, connectionName, err)
		return err
	}

	// update state table (inside transaction)
	err = tableUpdater.onConnectionReady(ctx, tx, connectionName)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to update connection state table")
	}
	return nil
}

func executeDeleteQueries(ctx context.Context, pool *pgxpool.Pool, deletions []string, tableUpdater *connectionStateTableUpdater) error {
	statushooks.SetStatus(ctx, fmt.Sprintf("Deleting %d %s", len(deletions), utils.Pluralize("connection", len(deletions))))

	var errors []error
	for _, c := range deletions {
		utils.LogTime("delete connection start")
		log.Printf("[TRACE] delete connection %s\n ", c)

		err := executeDeleteQuery(ctx, pool, tableUpdater, c)
		if err != nil {
			errors = append(errors, err)
		}
		utils.LogTime("delete connection end")
	}

	return error_helpers.CombineErrors(errors...)
}

func executeDeleteQuery(ctx context.Context, pool *pgxpool.Pool, tableUpdater *connectionStateTableUpdater, connectionName string) (err error) {
	sql := getDeleteConnectionQuery(connectionName)
	// create a transaction
	tx, err := pool.Begin(ctx)
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

	// execute delete sql
	_, err = tx.Exec(ctx, sql)
	if err != nil {
		tableUpdater.onConnectionError(ctx, tx, connectionName, err)
		return err
	}

	// delete state table (inside transaction)
	err = tableUpdater.onConnectionDeleted(ctx, tx, connectionName)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to delete connection state table")
	}
	return nil
}

func cloneConnectionSchemas(ctx context.Context, pool *pgxpool.Pool, pluginMap map[string]string, cloneableConnections steampipeconfig.ConnectionDataMap, idx int, numUpdates int, tableUpdater *connectionStateTableUpdater) error {
	var wg sync.WaitGroup
	var progressChan = make(chan string)
	type connectionError struct {
		name string
		err  error
	}
	var errChan = make(chan connectionError)

	var pluginMapMut sync.Mutex

	sem := semaphore.NewWeighted(int64(pool.Config().MaxConns))
	var errors []error

	go func() {
		for {
			select {
			case connectionError := <-errChan:
				errors = append(errors, connectionError.err)
				tableUpdater.onConnectionError(ctx, nil, connectionError.name, connectionError.err)
			case connectionName := <-progressChan:
				if connectionName == "" {
					return
				}
				idx++

			}
		}
	}()
	for n, d := range cloneableConnections {
		wg.Add(1)
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		// use semaphore to limit goroutines
		go func(connectionName string, connectionData *steampipeconfig.ConnectionData) {
			//log.Printf("[WARN] start clone connection %s", connectionName)
			defer func() {
				wg.Done()
				sem.Release(1)
			}()

			// this schema is already in the plugin map, clone from it
			exemplarSchemaName := pluginMap[connectionData.Plugin]

			// Clone the foreign schema into this connection.
			sql := fmt.Sprintf("select clone_foreign_schema('%s', '%s', '%s');", exemplarSchemaName, connectionName, connectionData.Plugin)
			// execute clone query, and update the connection state table, in a transaction
			if err := executeUpdateQuery(ctx, pool, tableUpdater, sql, connectionName); err != nil {
				errChan <- connectionError{connectionName, err}
				return
			}

			pluginMapMut.Lock()
			pluginMap[connectionData.Plugin] = connectionName
			pluginMapMut.Unlock()

			progressChan <- connectionName
		}(n, d)

	}

	wg.Wait()
	close(progressChan)

	return error_helpers.CombineErrors(errors...)
}

func getCommentsQueryForPlugin(connectionName string, p *steampipeconfig.ConnectionPlugin) string {
	var statements strings.Builder
	for t, schema := range p.ConnectionMap[connectionName].Schema.Schema {
		table := db_common.PgEscapeName(t)
		schemaName := db_common.PgEscapeName(connectionName)
		if schema.Description != "" {
			tableDescription := db_common.PgEscapeString(schema.Description)
			statements.WriteString(fmt.Sprintf("COMMENT ON FOREIGN TABLE %s.%s is %s;\n", schemaName, table, tableDescription))
		}
		for _, c := range schema.Columns {
			if c.Description != "" {
				column := db_common.PgEscapeName(c.Name)
				columnDescription := db_common.PgEscapeString(c.Description)
				statements.WriteString(fmt.Sprintf("COMMENT ON COLUMN %s.%s.%s is %s;\n", schemaName, table, column, columnDescription))
			}
		}
	}
	return statements.String()
}

func getUpdateConnectionQuery(localSchema, remoteSchema string) string {
	// escape the name
	localSchema = db_common.PgEscapeName(localSchema)

	var statements strings.Builder

	// Each connection has a unique schema. The schema, and all objects inside it,
	// are owned by the root user.
	statements.WriteString(fmt.Sprintf("drop schema if exists %s cascade;\n", localSchema))
	statements.WriteString(fmt.Sprintf("create schema %s;\n", localSchema))
	statements.WriteString(fmt.Sprintf("comment on schema %s is 'steampipe plugin: %s';\n", localSchema, remoteSchema))

	// Steampipe users are allowed to use the new schema
	statements.WriteString(fmt.Sprintf("grant usage on schema %s to steampipe_users;\n", localSchema))

	// Permissions are limited to select only, and should be granted for all new
	// objects. Steampipe users cannot create tables or modify data in the
	// connection schema - they need to use the public schema for that.  These
	// commands alter the defaults for any objects created in the future.
	// See https://www.postgresql.org/docs/12/ddl-priv.html
	statements.WriteString(fmt.Sprintf("alter default privileges in schema %s grant select on tables to steampipe_users;\n", localSchema))

	// If there are any objects already then grant their permissions now. (This
	// should not actually do anything at this point.)
	statements.WriteString(fmt.Sprintf("grant select on all tables in schema %s to steampipe_users;\n", localSchema))

	// Import the foreign schema into this connection.
	statements.WriteString(fmt.Sprintf("import foreign schema \"%s\" from server steampipe into %s;\n", remoteSchema, localSchema))

	return statements.String()
}

func getDeleteConnectionQuery(name string) string {
	return fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;\n", db_common.PgEscapeName(name))
}
