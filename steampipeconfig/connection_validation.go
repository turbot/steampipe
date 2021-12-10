package steampipeconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	sdkversion "github.com/turbot/steampipe-plugin-sdk/version"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
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

func ValidatePlugins(updates ConnectionDataMap, plugins map[string]*ConnectionPlugin) ([]*ValidationFailure, ConnectionDataMap, []*ConnectionPlugin) {
	var validatedPlugins []*ConnectionPlugin
	var validatedUpdates = ConnectionDataMap{}

	var validationFailures []*ValidationFailure
	for _, p := range plugins {
		if validationFailure := validateColumnDefVersion(p); validationFailure != nil {
			// validation failed
			validationFailures = append(validationFailures, validationFailure)
		} else if validationFailure := validateConnectionName(p); validationFailure != nil {
			// validation failed
			validationFailures = append(validationFailures, validationFailure)
		} else {
			// validation passed - add to list of validated plugins
			validatedPlugins = append(validatedPlugins, p)
			// if this connection has updates, add them
			if _, ok := updates[p.ConnectionName]; ok {
				validatedUpdates[p.ConnectionName] = updates[p.ConnectionName]
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
%s:

%s

%d %s was not imported.
`,
		constants.Red("Validation Errors"),
		strings.Join(warningsStrings, "\n\n"),
		failureCount,
		utils.Pluralize("connection", failureCount))
	return str
}

func validateConnectionName(p *ConnectionPlugin) *ValidationFailure {
	if helpers.StringSliceContains(constants.ReservedConnectionNames, p.ConnectionName) {
		return &ValidationFailure{
			Plugin:             p.PluginName,
			ConnectionName:     p.ConnectionName,
			Message:            fmt.Sprintf("Connection name cannot be one of %s", strings.Join(constants.ReservedConnectionNames, ",")),
			ShouldDropIfExists: false,
		}
	}
	return nil
}

func validateColumnDefVersion(p *ConnectionPlugin) *ValidationFailure {
	pluginProtocolVersion := p.Schema.GetProtocolVersion()
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
			ConnectionName:     p.ConnectionName,
			Message:            "Incompatible steampipe-plugin-sdk version. Please upgrade Steampipe.",
			ShouldDropIfExists: true,
		}
	}
	return nil
}
