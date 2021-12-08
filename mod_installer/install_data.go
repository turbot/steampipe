package mod_installer

import (
	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/version"
)

type InstallData struct {
	// record of the full dependency tree
	Lock modconfig.WorkspaceLock

	// all installed mod versions (some of which may not be required)
	AllInstalled modconfig.VersionListMap
	// mod versions which are installed but not needed
	//Unreferenced VersionListMap
	// ALL the available versions for each dependency mod(we populate this in a lazy fashion)
	AllAvailable modconfig.VersionListMap

	// list of dependencies installed by recent install operation
	RecentlyInstalled modconfig.ResolvedVersionListMap
	// list of dependencies which were already installed
	AlreadyInstalled modconfig.ResolvedVersionListMap
}

func NewInstallData(installedMods modconfig.VersionListMap, workspaceLock modconfig.WorkspaceLock) *InstallData {
	return &InstallData{
		Lock:         workspaceLock,
		AllInstalled: installedMods,
		//Unreferenced: make(VersionListMap),
		AllAvailable:      make(modconfig.VersionListMap),
		RecentlyInstalled: make(modconfig.ResolvedVersionListMap),
		AlreadyInstalled:  make(modconfig.ResolvedVersionListMap),
	}
}

// GetAvailableUpdates returns a map of all installed mods which are not in the lock file
func (s *InstallData) GetAvailableUpdates() (modconfig.WorkspaceLock, error) {
	res := make(modconfig.WorkspaceLock)
	for parent, deps := range s.Lock {
		for name, dep := range deps {
			availableVersions, err := s.getAvailableModVersions(name)
			if err != nil {
				return nil, err
			}
			constraint, _ := version.NewConstraint(dep.Constraint)
			var latestVersion = getVersionSatisfyingConstraint(constraint, availableVersions)
			if latestVersion.GreaterThan(dep.Version) {
				res.Add(name, latestVersion, constraint, parent)
			}
		}
	}
	return res, nil
}

// onModInstalled is called when a dependency is satisfied by installing a mod version
func (s *InstallData) onModInstalled(dependency *ResolvedModRef, parent *modconfig.Mod) {
	// update lock
	// get the constraint from the parent (it must be there)
	modVersion := parent.Requires.GetModDependency(dependency.Name)
	s.Lock.Add(dependency.Name, dependency.Version, modVersion.Constraint, parent.Name())
	// update list of all installed mods
	s.AllInstalled.Add(dependency.Name, dependency.Version)
	// update list items installed by this installer
	s.RecentlyInstalled.Add(dependency.Name, &modconfig.ResolvedVersionConstraint{
		Version:    dependency.Version,
		Constraint: dependency.Constraint.Original,
	})
}

// addExisting is called when a dependency is satisfied by a mod which is already installed
func (s *InstallData) addExisting(name string, version *semver.Version, parent *modconfig.Mod) {
	// update lock
	modVersion := parent.Requires.GetModDependency(name)
	s.Lock.Add(name, version, modVersion.Constraint, parent.Name())
	// update list of already installed items
	s.AlreadyInstalled.Add(name, &modconfig.ResolvedVersionConstraint{
		Version:    version,
		Constraint: modVersion.Constraint.Original,
	})
}

// return a map of all installed mods which are not in the lock file
func (s *InstallData) getUnusedMods() modconfig.VersionListMap {
	var unusedModPaths = make(modconfig.VersionListMap)
	// now delete any mod folders which are not in the lock file
	for name, versions := range s.AllInstalled {
		for _, version := range versions {
			if !s.Lock.ContainsModVersion(name, version) {
				unusedModPaths.Add(name, version)
			}
		}
	}
	return unusedModPaths
}

// retrieve all available mod versions from our cache, or from Git if not yet cached
func (s *InstallData) getAvailableModVersions(modName string) ([]*semver.Version, error) {
	// have we already loaded the versions for this mod
	availableVersions, ok := s.AllAvailable[modName]
	if ok {
		return availableVersions, nil
	}
	// so we have not cached this yet - retrieve from Git
	var err error
	availableVersions, err = getTagVersionsFromGit(getGitUrl(modName))
	if err != nil {
		return nil, err
	}
	// update our cache
	s.AllAvailable[modName] = availableVersions

	return availableVersions, nil
}
