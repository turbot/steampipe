package filepaths

import "strings"

func GetPluginShortName(pluginName string) string {
	if strings.HasPrefix(pluginName, "steampipe-plugin-") {
		return strings.TrimPrefix(pluginName, "steampipe-plugin-")
	}
	return pluginName
}
func GetPluginLongName(pluginName string) string {
	if !strings.HasPrefix(pluginName, "steampipe-plugin-") {
		return "steampipe-plugin-" + pluginName
	}
	return pluginName
}
