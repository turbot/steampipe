package db

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gertd/go-pluralize"
	"github.com/hashicorp/go-version"
	sdkversion "github.com/turbot/steampipe-plugin-sdk/version"
	"github.com/turbot/steampipe/connection_config"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

type validationFailure struct {
	plugin         string
	connectionName string
	message        string
}

func (v validationFailure) String() string {
	return fmt.Sprintf("Connection %s, Plugin %s: %s", v.connectionName, v.plugin, v.message)
}

func refreshConnections(client *Client) error {
	// first get a list of all existing schemas
	schemas := client.schemaMetadata.GetSchemas()

	// refresh the connection state file - the removes any connections which do not exist in the list of current schema
	log.Println("[TRACE] refreshConnections")
	updates, err := connection_config.GetConnectionsToUpdate(schemas)
	if err != nil {
		return err
	}
	log.Printf("[TRACE] updates: %+v\n", updates)

	missingCount := len(updates.MissingPlugins)
	if missingCount > 0 {
		// if any plugins are missing, error for now but we could prompt for an install
		p := pluralize.NewClient()
		return fmt.Errorf("%d %s referenced in the connection config not installed: \n  %v",
			missingCount,
			p.Pluralize("plugin", missingCount, false),
			strings.Join(updates.MissingPlugins, "\n  "))
	}

	var connectionQueries []string
	var warningString string
	numUpdates := len(updates.Update)
	if numUpdates > 0 {
		s := utils.ShowSpinner("Refreshing connections...")
		defer func() {
			utils.StopSpinner(s)
			// if any warnings were returned, display them on stderr
			if len(warningString) > 0 {
				// println writes to stderr
				println(constants.Red(warningString))
			}
		}()

		// first instantiate connection plugins for all updates
		connectionPlugins, err := getConnectionPlugins(updates.Update)
		if err != nil {
			return err
		}
		// find any plugins which use a newer sdk version than steampipe.
		validationFailures, validatedUpdates, validatedPlugins := validatePlugins(updates.Update, connectionPlugins)
		warningString = buildValidationWarningString(validationFailures)

		// get schema queries - this updates schemas for validated plugins and drops schemas for unvalidated plugins
		connectionQueries = getSchemaQueries(validatedUpdates, validationFailures)
		// add comments queries for validated connections
		connectionQueries = append(connectionQueries, getCommentQueries(validatedPlugins)...)
	}

	for c := range updates.Delete {
		log.Printf("[TRACE] delete %s\n ", c)
		connectionQueries = append(connectionQueries, deleteConnectionQuery(c)...)
	}
	if len(connectionQueries) > 0 {
		if err = executeConnectionQueries(connectionQueries, updates); err != nil {
			return err
		}
	} else {
		log.Println("[DEBUG] no connections to update")
	}

	return updateConnectionMapAndSchema(client)
}

func getConnectionPlugins(updates connection_config.ConnectionMap) ([]*connection_config.ConnectionPlugin, error) {
	var connectionPlugins []*connection_config.ConnectionPlugin
	numUpdates := len(updates)
	var pluginChan = make(chan *connection_config.ConnectionPlugin, numUpdates)
	var errorChan = make(chan error, numUpdates)
	for connectionName, plugin := range updates {
		pluginFQN := plugin.Plugin
		connectionConfig := plugin.ConnectionConfig

		// instantiate the connection plugin, and retrieve schema
		go getConnectionPluginsAsync(pluginFQN, connectionName, connectionConfig, pluginChan, errorChan)
	}

	for i := 0; i < numUpdates; i++ {
		select {
		case err := <-errorChan:
			log.Println("[TRACE] hydrate err chan select", "error", err)
			return nil, err
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("timed out retrieving schema from plugins")
		case p := <-pluginChan:
			connectionPlugins = append(connectionPlugins, p)
		}
	}

	return connectionPlugins, nil
}

func getConnectionPluginsAsync(pluginFQN string, connectionName string, connectionConfig string, pluginChan chan *connection_config.ConnectionPlugin, errorChan chan error) {
	opts := &connection_config.ConnectionPluginOptions{
		PluginFQN:        pluginFQN,
		ConnectionName:   connectionName,
		ConnectionConfig: connectionConfig,
		DisableLogger:    true}
	p, err := connection_config.CreateConnectionPlugin(opts)
	if err != nil {
		errorChan <- err
		return
	}
	pluginChan <- p

	p.Plugin.Client.Kill()
}

func validatePlugins(updates connection_config.ConnectionMap, plugins []*connection_config.ConnectionPlugin) ([]*validationFailure, connection_config.ConnectionMap, []*connection_config.ConnectionPlugin) {
	var validatedPlugins []*connection_config.ConnectionPlugin
	var validatedUpdates = connection_config.ConnectionMap{}

	var validationFailures []*validationFailure
	for _, p := range plugins {
		if validationFailure := validateSdkVersion(p); validationFailure != nil {
			// validation failed
			validationFailures = append(validationFailures, validationFailure)
		} else {
			// validation passed - add to liost of validated plugins
			validatedPlugins = append(validatedPlugins, p)
			validatedUpdates[p.ConnectionName] = updates[p.ConnectionName]
		}
	}
	return validationFailures, validatedUpdates, validatedPlugins

}

func validateSdkVersion(p *connection_config.ConnectionPlugin) *validationFailure {
	pluginSdkVersionString := p.Schema.SdkVersion
	if pluginSdkVersionString == "" {
		// plugins compiled against 0.1.x of the sdk do not return the version
		return nil
	}
	pluginSdkVersion, err := version.NewSemver(pluginSdkVersionString)
	if err != nil {
		return &validationFailure{
			plugin:         p.PluginName,
			connectionName: p.ConnectionName,
			message:        fmt.Sprintf("could not parse plugin sdk version %s", pluginSdkVersion),
		}
	}
	steampipeSdkVersion := sdkversion.SemVer
	if pluginSdkVersion.GreaterThan(steampipeSdkVersion) {
		return &validationFailure{
			plugin:         p.PluginName,
			connectionName: p.ConnectionName,
			message:        "plugin uses a more recent version of the steampipe-plugin-sdk than Steampipe",
		}
	}
	return nil
}

func buildValidationWarningString(failures []*validationFailure) string {
	if len(failures) == 0 {
		return ""
	}
	warningsStrings := []string{}
	for _, failure := range failures {
		warningsStrings = append(warningsStrings, failure.String())
	}
	p := pluralize.NewClient()
	failureCount := len(failures)
	str := fmt.Sprintf("\nPlugin validation errors - %d %s will not be imported:\n   %s \nPlease update Steampipe.\n", failureCount, p.Pluralize("connection", failureCount, false), strings.Join(warningsStrings, "\n   "))
	return str
}

func getSchemaQueries(updates connection_config.ConnectionMap, failures []*validationFailure) []string {
	var schemaQueries []string
	for connectionName, plugin := range updates {
		remoteSchema := connection_config.PluginFQNToSchemaName(plugin.Plugin)
		log.Printf("[TRACE] update connection %s, plugin FQN %s, schema %s\n ", connectionName, plugin.Plugin, remoteSchema)
		schemaQueries = append(schemaQueries, updateConnectionQuery(connectionName, remoteSchema)...)
	}
	for _, failure := range failures {
		log.Printf("[TRACE] remove schema for conneciton failing validation connection %s, plugin FQN %s\n ", failure.connectionName, failure.plugin)
		schemaQueries = append(schemaQueries, deleteConnectionQuery(failure.connectionName)...)
	}

	return schemaQueries
}

func getCommentQueries(plugins []*connection_config.ConnectionPlugin) []string {
	var commentQueries []string
	for _, plugin := range plugins {
		commentQueries = append(commentQueries, commentsQuery(plugin)...)
	}
	return commentQueries
}

func updateConnectionQuery(localSchema, remoteSchema string) []string {
	// escape the name
	localSchema = PgEscapeName(localSchema)
	return []string{

		// Each connection has a unique schema. The schema, and all objects inside it,
		// are owned by the root user.
		fmt.Sprintf(`drop schema if exists %s cascade;`, localSchema),
		fmt.Sprintf(`create schema %s;`, localSchema),
		fmt.Sprintf(`comment on schema %s is 'steampipe plugin: %s';`, localSchema, remoteSchema),

		// Steampipe users are allowed to use the new schema
		fmt.Sprintf(`grant usage on schema %s to steampipe_users;`, localSchema),

		// Permissions are limited to select only, and should be granted for all new
		// objects. Steampipe users cannot create tables or modify data in the
		// connection schema - they need to use the public schema for that.  These
		// commands alter the defaults for any objects created in the future.
		// See https://www.postgresql.org/docs/12/ddl-priv.html
		fmt.Sprintf(`alter default privileges in schema %s grant select on tables to steampipe_users;`, localSchema),

		// If there are any objects already then grant their permissions now. (This
		// should not actually do anything at this point.)
		fmt.Sprintf(`grant select on all tables in schema %s to steampipe_users;`, localSchema),

		// Import the foreign schema into this connection.
		fmt.Sprintf(`import foreign schema "%s" from server steampipe into %s;`, remoteSchema, localSchema),
	}
}

func commentsQuery(p *connection_config.ConnectionPlugin) []string {
	var statements []string
	for t, schema := range p.Schema.Schema {
		table := PgEscapeName(t)
		schemaName := PgEscapeName(p.ConnectionName)
		if schema.Description != "" {
			tableDescription := PgEscapeString(schema.Description)
			statements = append(statements, fmt.Sprintf("COMMENT ON FOREIGN TABLE %s.%s is %s;", schemaName, table, tableDescription))
		}
		for _, c := range schema.Columns {
			if c.Description != "" {
				column := PgEscapeName(c.Name)
				columnDescription := PgEscapeString(c.Description)
				statements = append(statements, fmt.Sprintf("COMMENT ON COLUMN %s.%s.%s is %s;", schemaName, table, column, columnDescription))
			}
		}
	}
	return statements
}

func deleteConnectionQuery(name string) []string {
	return []string{
		fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE;`, name),
	}
}

func executeConnectionQueries(schemaQueries []string, updates *connection_config.ConnectionUpdates) error {
	client, err := createSteampipeRootDbClient()
	if err != nil {
		return err
	}
	defer func() {
		client.Close()
	}()

	// combine queries
	schemaQueryString := strings.Join(schemaQueries, "\n")

	log.Printf("[DEBUG] there are connections to update, query: \n%s\n", schemaQueryString)
	_, err = client.Exec(schemaQueryString)
	if err != nil {
		return err
	}
	/* TODO - Log results
	log.Println("[TRACE] refresh connection results")
	for row := range schemaResult {
		log.Printf("[TRACE] %v\n", row)
	}
	*/

	// now update the state file
	err = connection_config.SaveConnectionState(updates.RequiredConnections)
	if err != nil {
		return err
	}
	return nil
}

func updateConnectionMapAndSchema(client *Client) error {
	// reload the database schemas, since they have changed
	// otherwise we wouldn't be here
	log.Println("[TRACE] reloading schema")
	client.loadSchema()

	// set the search path with the updates
	log.Println("[TRACE] setting search path")
	client.setSearchPath()

	// load the connection state and cache it!
	log.Println("[TRACE]", "retrieving connection map")
	connectionMap, err := connection_config.GetConnectionState(clientSingleton.schemaMetadata.GetSchemas())
	if err != nil {
		return err
	}
	log.Println("[TRACE]", "setting connection map")
	client.connectionMap = &connectionMap

	return nil
}
