package grpc

import (
	"strings"

	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

// HandleStartFailure is used to handle errors when starting both Steampipe plugins an dthe plugin manage
// (which is itself a GRPC plugin)
//
// When starting a GRPC plugin, a specific handshake sequence is expected on stdout.
// (This is automatically written in the case of a successfulty startup)
// If the handshae is missing (because the startup failed or anything else was written to stdout)
// we get the error "Unrecognized remote plugin message"
//
// If the plugin startup fails with an error panic, it constructs a message string
// starting with the prefix  "Plugin startup failed: " , detailing the error.
//
// This function checks whether the error returned from startup is "Unrecognized remote plugin message",
// and if so, it looks for ""Plugin startup failed: " in the plugin message and if found,
// extracts the underlying error message. This is returnerd as an error
func HandleStartFailure(err error) error {
	// extract the plugin message
	_, pluginMessage, found := strings.Cut(err.Error(), sdkplugin.UnrecognizedRemotePluginMessage)
	if !found {
		return err
	}
	pluginMessage, _, found = strings.Cut(pluginMessage, sdkplugin.UnrecognizedRemotePluginMessageSuffix)
	if !found {
		return err
	}

	// if this was an error during startup, reraise an error with the error string
	_, pluginError, found := strings.Cut(pluginMessage, sdkplugin.PluginStartupFailureMessage)
	if !found {
		return err
	}

	if strings.Contains(pluginMessage, sdkplugin.PluginStartupFailureMessage) {
		return sperr.New("%s", pluginError)
	}
	return err
}
