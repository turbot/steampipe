package connection

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	sdkversion "github.com/turbot/steampipe-plugin-sdk/v5/version"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

type ValidationFailure struct {
	Plugin             string
	ConnectionName     string
	Message            string
	ShouldDropIfExists bool
}

func (v ValidationFailure) String() string {
	return fmt.Sprintf(
		"Connection: %s\nPlugin:     %s\nError:      %s",
		v.ConnectionName,
		v.Plugin,
		v.Message,
	)
}

func ValidatePlugins(plugins map[string]*steampipeconfig.ConnectionPlugin) ([]*ValidationFailure, map[string]*steampipeconfig.ConnectionPlugin) {
	var validatedPlugins = make(map[string]*steampipeconfig.ConnectionPlugin)

	var validationFailures []*ValidationFailure
	for connectionName, connectionPlugin := range plugins {
		if validationFailure := validateProtocolVersion(connectionName, connectionPlugin); validationFailure != nil {
			// validation failed
			validationFailures = append(validationFailures, validationFailure)
		} else if validationFailure := validateConnectionName(connectionName, connectionPlugin); validationFailure != nil {
			// validation failed
			validationFailures = append(validationFailures, validationFailure)
		} else {
			validatedPlugins[connectionName] = connectionPlugin
		}
	}

	return validationFailures, validatedPlugins

}
func ValidateUpdates(updates, commentUpdates steampipeconfig.ConnectionStateMap, validatedPlugins map[string]*steampipeconfig.ConnectionPlugin) ([]*ValidationFailure, steampipeconfig.ConnectionStateMap, steampipeconfig.ConnectionStateMap) {
	var validatedUpdates = steampipeconfig.ConnectionStateMap{}
	var validatedCommentUpdates = steampipeconfig.ConnectionStateMap{}

	var validationFailures []*ValidationFailure
	for connectionName, _ := range validatedPlugins {
		// if this connection has updates, add them
		if _, ok := updates[connectionName]; ok {
			validatedUpdates[connectionName] = updates[connectionName]
		}
		// if this connection has comment updates, add them
		if _, ok := commentUpdates[connectionName]; ok {
			validatedCommentUpdates[connectionName] = validatedCommentUpdates[connectionName]
		}

	}

	// we need to separately validate aggregator connections as there will not be a connection plugin for them
	for connectionName, connectionState := range updates {
		if validateAggregator(connectionState, validatedPlugins) {
			validatedUpdates[connectionName] = connectionState
		}
	}
	return validationFailures, validatedUpdates, validatedCommentUpdates

}

func validateAggregator(connectionState *steampipeconfig.ConnectionState, validatedPlugins map[string]*steampipeconfig.ConnectionPlugin) bool {
	connectionName := connectionState.ConnectionName
	if connectionState.GetType() == modconfig.ConnectionTypeAggregator {
		// get the conneciton object
		connection := steampipeconfig.GlobalConfig.Connections[connectionName]
		// get the first child connection
		for _, childConnection := range connection.Connections {
			// check whether the plugin for this connection is validated
			for _, p := range validatedPlugins {
				return p.IncludesConnection(childConnection.Name)
			}
		}
	}
	return false
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

func validateConnectionName(connectionName string, p *steampipeconfig.ConnectionPlugin) *ValidationFailure {
	if helpers.StringSliceContains(constants.ReservedConnectionNames, connectionName) {
		return &ValidationFailure{
			Plugin:         p.PluginName,
			ConnectionName: connectionName,
			Message:        fmt.Sprintf("Connection name cannot be one of %s", strings.Join(constants.ReservedConnectionNames, ",")),
			// no need to drop - this connection cannot have been created as a schema
			// - we DO NOT want to drop one of the reserved schemas!
			ShouldDropIfExists: false,
		}
	}
	return nil
}

func validateProtocolVersion(connectionName string, p *steampipeconfig.ConnectionPlugin) *ValidationFailure {
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
			Message:        "Incompatible steampipe-plugin-sdk version. Please upgrade Steampipe.",
			// drop this connection if it exists
			ShouldDropIfExists: true,
		}
	}
	return nil
}
