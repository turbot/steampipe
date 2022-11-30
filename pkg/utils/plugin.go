package utils

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
)

const maxSchemaNameLength = 63

// PluginFQNToSchemaName convert a full plugin name to a schema name
// schemas in postgres are limited to 63 chars - the name may be longer than this, in which case trim the length
// and add a hash to the end to make unique
func PluginFQNToSchemaName(pluginFQN string) string {
	if len(pluginFQN) < maxSchemaNameLength {
		return pluginFQN
	}

	schemaName := TrimSchemaName(pluginFQN) + fmt.Sprintf("-%x", helpers.StringFnvHash(pluginFQN))
	return schemaName
}

func TrimSchemaName(pluginFQN string) string {
	if len(pluginFQN) < maxSchemaNameLength {
		return pluginFQN
	}

	return pluginFQN[:maxSchemaNameLength-9]
}

// GetPluginName function is used to get the plugin name required while
// installing/updating/removing a plugin. External plugins require the repo
// names to be prefixed(eg: francois2metz/scalingo).
// Sample input 1: hub.steampipe.io/plugins/francois2metz/scalingo@latest
// Sample output 1: francois2metz/scalingo
// Sample input 2: hub.steampipe.io/plugins/turbot/aws@latest
// Sample output 2: aws
func GetPluginName(plugin string) string {
	repo := strings.Split(plugin, "/")[2]
	p := strings.Split(plugin, "/")[3]
	plugin_name := strings.Split(p, "@")[0]

	if repo == "turbot" {
		return plugin_name
	}

	return fmt.Sprintf("%s/%s", repo, plugin_name)
}
