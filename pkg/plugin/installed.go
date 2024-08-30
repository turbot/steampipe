package plugin

import (
	"context"
	"fmt"
	"github.com/turbot/pipe-fittings/ociinstaller"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// GetInstalledPlugins returns the list of plugins keyed by the shortname (org/name) and its specific version
// Does not validate/check of available connections
func GetInstalledPlugins(ctx context.Context) (map[string]*modconfig.PluginVersionString, error) {
	installedPlugins := make(map[string]*modconfig.PluginVersionString)
	installedPluginsData, _ := List(ctx, nil)
	for _, plugin := range installedPluginsData {
		org, name, _ := ociinstaller.NewImageRef(plugin.Name).GetOrgNameAndConstraint(constants.SteampipeHubOCIBase)
		pluginShortName := fmt.Sprintf("%s/%s", org, name)
		installedPlugins[pluginShortName] = plugin.Version
	}
	return installedPlugins, nil
}
