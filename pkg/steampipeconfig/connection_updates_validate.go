package steampipeconfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/utils"
	sdkversion "github.com/turbot/steampipe-plugin-sdk/v5/version"
)

func (u *ConnectionUpdates) validate() {
	// find any plugins which use a newer sdk version than steampipe, and any connections with an invalid name
	u.validatePluginsAndConnections()
	u.validateUpdates()
}

func (u *ConnectionUpdates) validatePluginsAndConnections() {
	// TODO should plugin manager do this when starting the plugin???
	var validatedPlugins = make(map[string]*ConnectionPlugin)

	for connectionName, connectionPlugin := range u.ConnectionPlugins {
		if validationFailure := validateProtocolVersion(connectionName, connectionPlugin); validationFailure != nil {
			u.InvalidConnections[connectionName] = validationFailure
		} else if validationFailure := validateConnectionName(connectionName, connectionPlugin); validationFailure != nil {
			u.InvalidConnections[connectionName] = validationFailure
		} else {
			validatedPlugins[connectionName] = connectionPlugin
		}
	}

	// update connection plugins to only include validated
	u.ConnectionPlugins = validatedPlugins
}

func (u *ConnectionUpdates) validateUpdates() {
	var validatedUpdates = ConnectionStateMap{}
	var validatedCommentUpdates = ConnectionStateMap{}

	// ConnectionPlugins has now been validated and only contains valid connection plugins
	// for every update and comment update, confirm the connection plugin is valid
	for connectionName, connectionState := range u.Update {
		if _, ok := u.ConnectionPlugins[connectionName]; ok {
			// if this connection has a validated connection plugin, add to valdiated updates
			validatedUpdates[connectionName] = connectionState
		} else {
			// try to get the validation failure - should be in InvalidConnections
			validationFailure, ok := u.InvalidConnections[connectionName]
			if ok {
				log.Printf("[WARN] validateUpdates - connection update '%s' failed validation: %s", connectionName, validationFailure.Message)
			} else {
				// not expected
				// for some reason there was no validation failure in the map
				log.Printf("[WARN] validateUpdates - connection update '%s' failed validation (connection not found in validated ConnectionPlugins but InvalidConnections does not contain the connection - this is unexpected)", connectionName)
			}
		}
	}

	for connectionName, connectionState := range u.MissingComments {
		// if this connection has a validated connection plugin, add to validated comment updates
		if _, ok := u.ConnectionPlugins[connectionName]; ok {
			validatedCommentUpdates[connectionName] = connectionState
		}
	}

	// now write back validated updates
	u.Update = validatedUpdates
	u.MissingComments = validatedCommentUpdates
}

func validateConnectionName(connectionName string, p *ConnectionPlugin) *ValidationFailure {
	if err := ValidateConnectionName(connectionName); err != nil {
		return &ValidationFailure{
			Plugin:         p.PluginName,
			ConnectionName: connectionName,
			Message:        err.Error(),
			// no need to drop - this connection cannot have been created as a schema
			ShouldDropIfExists: false,
		}
	}

	return nil
}

func validateProtocolVersion(connectionName string, p *ConnectionPlugin) *ValidationFailure {
	pluginProtocolVersion := p.ConnectionMap[connectionName].Schema.GetProtocolVersion()
	// if this is 0, the plugin does not define a protocol version
	// - so we know the plugin sdk version is older that the one we are using
	// therefore we are compatible
	if pluginProtocolVersion == 0 {
		return nil
	}

	steampipeProtocolVersion := sdkversion.ProtocolVersion
	if steampipeProtocolVersion < pluginProtocolVersion {
		return &ValidationFailure{
			Plugin:         p.PluginName,
			ConnectionName: connectionName,
			Message:        "Incompatible steampipe-plugin-sdk version. Please upgrade Steampipe to use this plugin.",
			// drop this connection if it exists
			ShouldDropIfExists: true,
		}
	}
	return nil
}

func BuildValidationWarningString(failures []*ValidationFailure) string {
	if len(failures) == 0 {
		return ""
	}
	warningsStrings := []string{}
	for _, failure := range failures {
		warningsStrings = append(warningsStrings, failure.String())
	}
	/*
		Plugin validation errors - 2 connections will not be imported, as they refer to plugins with a more recent version of the steampipe-plugin-sdk than Steampipe.
		   connection: gcp, plugin: hub.steampipe.io/plugins/turbot/gcp@latest
		   connection: aws, plugin: hub.steampipe.io/plugins/turbot/aws@latest
		Please update Steampipe in order to use these plugins
	*/
	failureCount := len(failures)
	str := fmt.Sprintf(`

%s

%s

%d %s not imported.
`,
		constants.Red(fmt.Sprintf("%d Connection Validation %s", failureCount, utils.Pluralize("Error", failureCount))),
		strings.Join(warningsStrings, "\n\n"),
		failureCount,
		utils.Pluralize("connection", failureCount))
	return str
}
