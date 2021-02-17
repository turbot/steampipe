package db

import (
	"fmt"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/hashicorp/go-version"
	sdkversion "github.com/turbot/steampipe-plugin-sdk/version"
	"github.com/turbot/steampipe/connection_config"
)

type validationFailure struct {
	plugin         string
	connectionName string
	message        string
}

func (v validationFailure) String() string {
	return fmt.Sprintf("  connection: %s\n  plugin: %s\n", v.connectionName, v.plugin)
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
	/*
		Plugin validation errors - 2 connections will not be imported, as they refer to plugins with a more recent version of the steampipe-plugin-sdk than Steampipe.
		   connection: gcp, plugin: hub.steampipe.io/plugins/turbot/gcp@latest
		   connection: aws, plugin: hub.steampipe.io/plugins/turbot/aws@latest
		Please update Steampipe in order to use these plugins
	*/
	p := pluralize.NewClient()
	failureCount := len(failures)
	p.AddPluralRule("this connection", "these connections")
	str := fmt.Sprintf("\nPlugin validation errors - %d %s will not be imported, as they refer to plugins with a more recent version of the steampipe-plugin-sdk than Steampipe.\n\n%s \nPlease update Steampipe in order to use %s.\n",
		failureCount,
		p.Pluralize("connection", failureCount, false),
		strings.Join(warningsStrings, "\n"),
		p.Pluralize("this connection", failureCount, false))
	return str
}
