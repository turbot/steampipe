package steampipeconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	sdkversion "github.com/turbot/steampipe-plugin-sdk/v5/version"
	"github.com/turbot/steampipe/pkg/constants"
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

func ValidatePlugins(updates ConnectionDataMap, plugins map[string]*ConnectionPlugin) ([]*ValidationFailure, ConnectionDataMap, map[string]*ConnectionPlugin) {
	var validatedPlugins = make(map[string]*ConnectionPlugin)
	var validatedUpdates = ConnectionDataMap{}

	var validationFailures []*ValidationFailure
	for connectionName, connectionPlugin := range plugins {
		if validationFailure := validateColumnDefVersion(connectionName, connectionPlugin); validationFailure != nil {
			// validation failed
			validationFailures = append(validationFailures, validationFailure)
		} else if validationFailure := validateConnectionName(connectionName, connectionPlugin); validationFailure != nil {
			// validation failed
			validationFailures = append(validationFailures, validationFailure)
		} else {
			// validation passed - add to list of validated plugins
			validatedPlugins[connectionName] = connectionPlugin
			// if this connection has updates, add them
			if _, ok := updates[connectionName]; ok {
				validatedUpdates[connectionName] = updates[connectionName]
			}
		}
	}

	// we need to separately validate aggregator connections as there will not be a connection plugin for them
	for updateConnectionName, connectionData := range updates {
		if connectionData.Connection.Type == modconfig.ConnectionTypeAggregator {
			// get the first child connection
			childConnection := connectionData.Connection.FirstChild()
			// check whether the plugin for this connection is validated
			for _, p := range validatedPlugins {
				if p.IncludesConnection(childConnection.Name) {
					validatedUpdates[updateConnectionName] = connectionData
				}
			}

		}
	}
	return validationFailures, validatedUpdates, validatedPlugins

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

func validateConnectionName(connectionName string, p *ConnectionPlugin) *ValidationFailure {
	if helpers.StringSliceContains(constants.ReservedConnectionNames, connectionName) {
		return &ValidationFailure{
			Plugin:             p.PluginName,
			ConnectionName:     connectionName,
			Message:            fmt.Sprintf("Connection name cannot be one of %s", strings.Join(constants.ReservedConnectionNames, ",")),
			ShouldDropIfExists: false,
		}
	}
	return nil
}

func validateColumnDefVersion(connectionName string, p *ConnectionPlugin) *ValidationFailure {
	pluginProtocolVersion := p.ConnectionMap[connectionName].Schema.GetProtocolVersion()
	// if this is 0, the plugin does not define columnDefinitionVersion
	// - so we know the plugin sdk version is older that the one we are using
	// therefore we are compatible
	if pluginProtocolVersion == 0 {
		return nil
	}

	steampipeProtocolVersion := sdkversion.ProtocolVersion
	if steampipeProtocolVersion < pluginProtocolVersion {
		return &ValidationFailure{
			Plugin:             p.PluginName,
			ConnectionName:     connectionName,
			Message:            "Incompatible steampipe-plugin-sdk version. Please upgrade Steampipe.",
			ShouldDropIfExists: true,
		}
	}
	return nil
}
