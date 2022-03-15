package modinstaller

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/versionmap"
	"github.com/turbot/steampipe/versionhelpers"
	"github.com/xlab/treeprint"
)

type InstallData struct {
	// record of the full dependency tree
	Lock    *versionmap.WorkspaceLock
	NewLock *versionmap.WorkspaceLock

	// ALL the available versions for each dependency mod (we populate this in a lazy fashion)
	allAvailable versionmap.VersionListMap

	// list of dependencies installed by recent install operation
	Installed versionmap.DependencyVersionMap
	// list of dependencies which have been upgraded
	Upgraded versionmap.DependencyVersionMap
	// list of dependencies which have been downgraded
	Downgraded versionmap.DependencyVersionMap
	// list of dependencies which have been uninstalled
	Uninstalled  versionmap.DependencyVersionMap
	WorkspaceMod *modconfig.Mod
}

func NewInstallData(workspaceLock *versionmap.WorkspaceLock, workspaceMod *modconfig.Mod) *InstallData {
	return &InstallData{
		Lock:         workspaceLock,
		WorkspaceMod: workspaceMod,
		NewLock:      versionmap.EmptyWorkspaceLock(workspaceLock),
		allAvailable: make(versionmap.VersionListMap),
		Installed:    make(versionmap.DependencyVersionMap),
		Upgraded:     make(versionmap.DependencyVersionMap),
		Downgraded:   make(versionmap.DependencyVersionMap),
		Uninstalled:  make(versionmap.DependencyVersionMap),
	}
}

// onModInstalled is called when a dependency is satisfied by installing a mod version
func (d *InstallData) onModInstalled(dependency *ResolvedModRef, parent *modconfig.Mod) {
	parentPath := parent.GetModDependencyPath()
	// get the constraint from the parent (it must be there)
	modVersion := parent.Require.GetModDependency(dependency.Name)
	// update lock
	d.NewLock.Add(dependency.Name, dependency.Version, modVersion.Constraint.Original, parentPath)
}

// addExisting is called when a dependency is satisfied by a mod which is already installed
func (d *InstallData) addExisting(name string, version *semver.Version, constraint *versionhelpers.Constraints, parent *modconfig.Mod) {
	// update lock
	parentPath := parent.GetModDependencyPath()
	d.NewLock.Add(name, version, constraint.Original, parentPath)
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

// update the lock with the NewLock and determine if any mods have been uninstalled
func (d *InstallData) onInstallComplete() {
	d.Installed = d.NewLock.GetMissingFromOther(d.Lock)
	d.Uninstalled = d.Lock.GetMissingFromOther(d.NewLock)
	d.Upgraded = d.Lock.GetUpgradedInOther(d.NewLock)
	d.Downgraded = d.Lock.GetDowngradedInOther(d.NewLock)
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

// GetLockedModVersion looks for a lock file entry matching the required constraint and returns nil if not found
// it checks both the existing and new lock files
func (d *InstallData) GetLockedModVersion(requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod) (*versionmap.ResolvedVersionConstraint, error) {
	// first try the d.Lock - this is the mods which were already installed before this instrallation started
	lockedVersion, err := d.Lock.GetLockedModVersion(requiredModVersion, parent)
	if lockedVersion != nil || err != nil {
		return lockedVersion, err
	}

	// if no version was found, try d.NewLock - this is mods which have been installed as part of this installation
	return d.NewLock.GetLockedModVersion(requiredModVersion, parent)
}
