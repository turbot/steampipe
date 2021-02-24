package connection_config

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/turbot/go-kit/helpers"
	sdkversion "github.com/turbot/steampipe-plugin-sdk/version"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

type ValidationFailure struct {
	Plugin         string
	ConnectionName string
	Message        string
}

func (v ValidationFailure) String() string {
	return fmt.Sprintf(
		"connection: %s\nplugin:     %s\nerror:      %s",
		constants.Bold(v.ConnectionName),
		constants.Bold(v.Plugin),
		constants.Bold(v.Message),
	)
}

func ValidatePlugins(updates ConnectionMap, plugins []*ConnectionPlugin) ([]*ValidationFailure, ConnectionMap, []*ConnectionPlugin) {
	var validatedPlugins []*ConnectionPlugin
	var validatedUpdates = ConnectionMap{}

	var validationFailures []*ValidationFailure
	for _, p := range plugins {
		if validationFailure := validateSdkVersion(p); validationFailure != nil {
			// validation failed
			validationFailures = append(validationFailures, validationFailure)
		} else if validationFailure := validateConnectionName(p); validationFailure != nil {
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
	str := fmt.Sprintf(`%s:

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
			Plugin:         p.PluginName,
			ConnectionName: p.ConnectionName,
			Message:        fmt.Sprintf("Connection name cannot be one of %s", strings.Join(constants.ReservedConnectionNames, ",")),
		}
	}
	return nil
}

func validateSdkVersion(p *ConnectionPlugin) *ValidationFailure {
	pluginSdkVersionString := p.Schema.SdkVersion
	if pluginSdkVersionString == "" {
		// plugins compiled against 0.1.x of the sdk do not return the version
		return nil
	}
	pluginSdkVersion, err := version.NewSemver(pluginSdkVersionString)
	if err != nil {
		return &ValidationFailure{
			Plugin:         p.PluginName,
			ConnectionName: p.ConnectionName,
			Message:        fmt.Sprintf("Could not parse plugin sdk version %s.", pluginSdkVersion),
		}
	}
	steampipeSdkVersion := sdkversion.SemVer
	if !validateIgnoringPrerelease(pluginSdkVersion, steampipeSdkVersion) {
		return &ValidationFailure{
			Plugin:         p.PluginName,
			ConnectionName: p.ConnectionName,
			Message:        "Incompatible steampipe-plugin-sdk version. Please upgrade Steampipe.",
		}
	}
	return nil
}

// return false if pluginSdkVersion is > steampipeSdkVersion, ignoring prerelease
func validateIgnoringPrerelease(pluginSdkVersion *version.Version, steampipeSdkVersion *version.Version) bool {
	pluginSegments := pluginSdkVersion.Segments()
	steampipeSegments := steampipeSdkVersion.Segments()
	return pluginSegments[0] <= steampipeSegments[0] && pluginSegments[1] <= steampipeSegments[1]

}
