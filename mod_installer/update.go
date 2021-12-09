package mod_installer

import "github.com/turbot/steampipe/steampipeconfig/version_map"

func GetAvailableUpdates(opts *InstallOpts) (installedMods version_map.DependencyVersionMap, availableUpdates version_map.DependencyVersionMap, err error) {
	// install workspace dependencies
	installer, err := NewModInstaller(opts)
	if err != nil {
		return nil, nil, err
	}
	availableUpdates, err = installer.installData.GetAvailableUpdates()
	if err != nil {
		return
	}
	installedMods = installer.installData.Lock.InstallCache
	return
}
