package filepaths

import "strings"

func PluginAliasToShortName(pluginName string) string {
	// remove  prefix
	split := strings.Split(pluginName, "/")
	pluginName = split[len(split)-1]

	if strings.HasPrefix(pluginName, "steampipe-plugin-") {
		return strings.TrimPrefix(pluginName, "steampipe-plugin-")
	}
	return pluginName
}
func PluginAliasToLongName(pluginName string) string {
	// remove  prefix
	split := strings.Split(pluginName, "/")
	pluginName = split[len(split)-1]

	if !strings.HasPrefix(pluginName, "steampipe-plugin-") {
		return "steampipe-plugin-" + pluginName
	}
	return pluginName
}
