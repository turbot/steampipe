package mod_installer

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/version_map"
	"github.com/turbot/steampipe/version"
)

type InstallData struct {
	// record of the full dependency tree
	Lock    *version_map.WorkspaceLock
	NewLock *version_map.WorkspaceLock

	// ALL the available versions for each dependency mod(we populate this in a lazy fashion)
	allAvailable version_map.VersionListMap

	// list of dependencies installed by recent install operation
	RecentlyInstalled version_map.ResolvedVersionListMap
	// list of dependencies which were already installed
	AlreadyInstalled version_map.ResolvedVersionListMap
	// list of dependencies which have been updated
	Updated version_map.ResolvedVersionListMap
	// list of dependencies which have been uninstalled
	Uninstalled version_map.ResolvedVersionListMap
}

func NewInstallData(workspaceLock *version_map.WorkspaceLock) *InstallData {
	return &InstallData{
		Lock:              workspaceLock,
		NewLock:           version_map.EmptyWorkspaceLock(workspaceLock),
		allAvailable:      make(version_map.VersionListMap),
		RecentlyInstalled: make(version_map.ResolvedVersionListMap),
		AlreadyInstalled:  make(version_map.ResolvedVersionListMap),
	}
}

// GetAvailableUpdates returns a map of all installed mods which are not in the lock file
func (d *InstallData) GetAvailableUpdates() (version_map.DependencyVersionMap, error) {
	res := make(version_map.DependencyVersionMap)
	for parent, deps := range d.Lock.InstallCache {
		for name, resolvedConstraint := range deps {
			availableVersions, err := d.getAvailableModVersions(name)
			if err != nil {
				return nil, err
			}
			constraint, _ := version.NewConstraint(resolvedConstraint.Constraint)
			var latestVersion = getVersionSatisfyingConstraint(constraint, availableVersions)
			if latestVersion.GreaterThan(resolvedConstraint.Version) {
				res.Add(name, latestVersion, constraint.Original, parent)
			}
		}
	}
	return res, nil
}

// onModInstalled is called when a dependency is satisfied by installing a mod version
func (d *InstallData) onModInstalled(dependency *ResolvedModRef, parent *modconfig.Mod) {
	// get the constraint from the parent (it must be there)
	modVersion := parent.Require.GetModDependency(dependency.Name)
	// update lock
	d.NewLock.InstallCache.Add(dependency.Name, dependency.Version, modVersion.Constraint.Original, parent.GetModDependencyPath())
	// update list items installed by this installer
	d.RecentlyInstalled.Add(dependency.Name, &version_map.ResolvedVersionConstraint{
		Name:       dependency.Name,
		Version:    dependency.Version,
		Constraint: dependency.Constraint.Original,
	})
}

// addExisting is called when a dependency is satisfied by a mod which is already installed
func (d *InstallData) addExisting(name string, version *semver.Version, constraint *version.Constraints, parent *modconfig.Mod) {
	// update lock
	d.NewLock.InstallCache.Add(name, version, constraint.Original, parent.GetModDependencyPath())

	modVersion := parent.Require.GetModDependency(name)
	// update list of already installed items
	d.AlreadyInstalled.Add(name, &version_map.ResolvedVersionConstraint{
		Name:       name,
		Version:    version,
		Constraint: modVersion.Constraint.Original,
	})
}

// retrieve all available mod versions from our cache, or from Git if not yet cached
func (d *InstallData) getAvailableModVersions(modName string) ([]*semver.Version, error) {
	// have we already loaded the versions for this mod
	availableVersions, ok := d.allAvailable[modName]
	if ok {
		return availableVersions, nil
	}
	// so we have not cached this yet - retrieve from Git
	var err error
	availableVersions, err = getTagVersionsFromGit(getGitUrl(modName))
	if err != nil {
		return nil, fmt.Errorf("could not retrieve version data from Git URL '%s'", modName)
	}
	// update our cache
	d.allAvailable[modName] = availableVersions

	return availableVersions, nil
}

// update the lock with the NewLock and dtermine if any mods have been uninstalled
func (d *InstallData) onInstallComplete() {
	d.Uninstalled = d.getUninstalled(d.Lock, d.NewLock)
	d.Lock = d.NewLock
}

// determine which dependencies have been uninstalled by comparing old and new lock data
func (d *InstallData) getUninstalled(oldLock *version_map.WorkspaceLock, newLock *version_map.WorkspaceLock) version_map.ResolvedVersionListMap {
	res := make(version_map.ResolvedVersionListMap)
	oldFlat := oldLock.InstallCache.FlatMap()
	newFlat := newLock.InstallCache.FlatMap()
	for fullName, oldDep := range oldFlat {
		if _, ok := newFlat[fullName]; !ok {
			res.Add(oldDep.Name, oldDep)
		}
	}
	return res
}
