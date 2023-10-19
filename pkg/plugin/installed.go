package plugin

import (
	"fmt"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/ociinstaller"
)

// GetInstalledPlugins returns the list of plugins keyed by the shortname (org/name) and its specific version
// Does not validate/check of available connections
func GetInstalledPlugins() (map[string]*modconfig.PluginVersionString, error) {
	installedPlugins := make(map[string]*modconfig.PluginVersionString)
	installedPluginsData, _ := List(nil)
	for _, plugin := range installedPluginsData {
		org, name, _ := ociinstaller.NewSteampipeImageRef(plugin.Name).GetOrgNameAndStream()
		pluginShortName := fmt.Sprintf("%s/%s", org, name)
		installedPlugins[pluginShortName] = plugin.Version
	}
	return installedPlugins, nil
}
