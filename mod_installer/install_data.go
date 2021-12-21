package mod_installer

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/version_map"
	"github.com/turbot/steampipe/version_helpers"
	"github.com/xlab/treeprint"
)

type InstallData struct {
	// record of the full dependency tree
	Lock    *version_map.WorkspaceLock
	NewLock *version_map.WorkspaceLock

	// ALL the available versions for each dependency mod(we populate this in a lazy fashion)
	allAvailable version_map.VersionListMap

	// list of dependencies installed by recent install operation
	Installed version_map.DependencyVersionMap
	// list of dependencies which have been upgraded
	Upgraded version_map.DependencyVersionMap
	// list of dependencies which have been downgraded
	Downgraded version_map.DependencyVersionMap
	// list of dependencies which have been uninstalled
	Uninstalled  version_map.DependencyVersionMap
	WorkspaceMod *modconfig.Mod
}

func NewInstallData(workspaceLock *version_map.WorkspaceLock, workspaceMod *modconfig.Mod) *InstallData {
	return &InstallData{
		Lock:         workspaceLock,
		WorkspaceMod: workspaceMod,
		NewLock:      version_map.EmptyWorkspaceLock(workspaceLock),
		allAvailable: make(version_map.VersionListMap),
		Installed:    make(version_map.DependencyVersionMap),
		Upgraded:     make(version_map.DependencyVersionMap),
		Downgraded:   make(version_map.DependencyVersionMap),
		Uninstalled:  make(version_map.DependencyVersionMap),
	}
}

// GetAvailableUpdates returns a map of all installed mods which are not in the lock file
func (d *InstallData) GetAvailableUpdates() (version_map.DependencyVersionMap, error) {
	res := make(version_map.DependencyVersionMap)
	for parent, deps := range d.Lock.InstallCache {
		for name, resolvedConstraint := range deps {
			includePrerelease := resolvedConstraint.IsPrerelease()
			availableVersions, err := d.getAvailableModVersions(name, includePrerelease)
			if err != nil {
				return nil, err
			}
			constraint, _ := version_helpers.NewConstraint(resolvedConstraint.Constraint)
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
	parentPath := parent.GetModDependencyPath()
	// get the constraint from the parent (it must be there)
	modVersion := parent.Require.GetModDependency(dependency.Name)
	// update lock
	d.NewLock.InstallCache.Add(dependency.Name, dependency.Version, modVersion.Constraint.Original, parentPath)
}

// addExisting is called when a dependency is satisfied by a mod which is already installed
func (d *InstallData) addExisting(name string, version *semver.Version, constraint *version_helpers.Constraints, parent *modconfig.Mod) {
	// update lock
	parentPath := parent.GetModDependencyPath()
	d.NewLock.InstallCache.Add(name, version, constraint.Original, parentPath)
}

// retrieve all available mod versions from our cache, or from Git if not yet cached
func (d *InstallData) getAvailableModVersions(modName string, includePrerelease bool) ([]*semver.Version, error) {
	// have we already loaded the versions for this mod
	availableVersions, ok := d.allAvailable[modName]
	if ok {
		return availableVersions, nil
	}
	// so we have not cached this yet - retrieve from Git
	var err error
	availableVersions, err = getTagVersionsFromGit(getGitUrl(modName), includePrerelease)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve version data from Git URL '%s'", modName)
	}
	// update our cache
	d.allAvailable[modName] = availableVersions

	return availableVersions, nil
}

// update the lock with the NewLock and dtermine if any mods have been uninstalled
func (d *InstallData) onInstallComplete() {
	d.Installed = d.NewLock.InstallCache.GetMissingFromOther(d.Lock.InstallCache)
	d.Uninstalled = d.Lock.InstallCache.GetMissingFromOther(d.NewLock.InstallCache)
	d.Upgraded = d.Lock.InstallCache.GetUpgradedInOther(d.NewLock.InstallCache)
	d.Downgraded = d.Lock.InstallCache.GetDowngradedInOther(d.NewLock.InstallCache)
	d.Lock = d.NewLock
}

func (d *InstallData) GetUpdatedTree() treeprint.Tree {
	return d.Upgraded.GetDependencyTree(d.WorkspaceMod.GetModDependencyPath())
}

func (d *InstallData) GetInstalledTree() treeprint.Tree {
	return d.Installed.GetDependencyTree(d.WorkspaceMod.GetModDependencyPath())
}

func (d *InstallData) GetUninstalledTree() treeprint.Tree {
	return d.Uninstalled.GetDependencyTree(d.WorkspaceMod.GetModDependencyPath())
}
