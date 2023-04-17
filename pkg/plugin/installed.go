package plugin

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/pkg/ociinstaller"
)

// GetInstalledPlugins returns the list of plugins keyed by the shortname (org/name) and its specific version
// Does not validate/check of available connections
func GetInstalledPlugins() (map[string]*semver.Version, error) {
	installedPlugins := make(map[string]*semver.Version)
	installedPluginsData, _ := List(nil)
	for _, plugin := range installedPluginsData {
		org, name, _ := ociinstaller.NewSteampipeImageRef(plugin.Name).GetOrgNameAndStream()
		semverVersion, err := semver.NewVersion(plugin.Version)
		if err != nil {
			continue
		}
		pluginShortName := fmt.Sprintf("%s/%s", org, name)
		installedPlugins[pluginShortName] = semverVersion
	}
	return installedPlugins, nil
}
