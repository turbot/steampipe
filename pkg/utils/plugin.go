package utils

import (
	"fmt"
	"strings"
)

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
