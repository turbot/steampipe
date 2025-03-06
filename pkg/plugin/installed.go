package plugin

import (
	"context"
	"fmt"

	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/versionfile"
)

// GetInstalledPlugins returns the list of plugins keyed by the shortname (org/name) and its specific version
// Does not validate/check of available connections
func GetInstalledPlugins(ctx context.Context, pluginVersions map[string]*versionfile.InstalledVersion) (map[string]*plugin.PluginVersionString, error) {
	installedPlugins := make(map[string]*plugin.PluginVersionString)
	installedPluginsData, _ := List(ctx, nil, pluginVersions)
	for _, plugin := range installedPluginsData {
		org, name, _ := ociinstaller.NewImageRef(plugin.Name).GetOrgNameAndStream()
		pluginShortName := fmt.Sprintf("%s/%s", org, name)
		installedPlugins[pluginShortName] = plugin.Version
	}
	return installedPlugins, nil
}
