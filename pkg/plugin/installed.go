package plugin

import (
	"context"
	"fmt"
	"github.com/turbot/pipe-fittings/ociinstaller"
	"github.com/turbot/pipe-fittings/ociinstaller/versionfile"
	plugin2 "github.com/turbot/pipe-fittings/plugin"
)

// GetInstalledPlugins returns the list of plugins keyed by the shortname (org/name) and its specific version
// Does not validate/check of available connections
func GetInstalledPlugins(ctx context.Context, pluginVersions map[string]*versionfile.InstalledVersion) (map[string]*plugin2.PluginVersionString, error) {
	installedPlugins := make(map[string]*plugin2.PluginVersionString)
	installedPluginsData, _ := List(ctx, nil, pluginVersions)
	for _, plugin := range installedPluginsData {
		org, name, _ := ociinstaller.NewImageRef(plugin.Name).GetOrgNameAndStream()
		pluginShortName := fmt.Sprintf("%s/%s", org, name)
		installedPlugins[pluginShortName] = plugin.Version
	}
	return installedPlugins, nil
}
