package mod_installer

import (
	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/version_map"
	"github.com/turbot/steampipe/version"
)

type InstallData struct {
	// record of the full dependency tree
	Lock *version_map.WorkspaceLock

	// ALL the available versions for each dependency mod(we populate this in a lazy fashion)
	allAvailable version_map.VersionListMap

	// list of dependencies installed by recent install operation
	RecentlyInstalled version_map.ResolvedVersionListMap
	// list of dependencies which were already installed
	AlreadyInstalled version_map.ResolvedVersionListMap
}

func NewInstallData(workspaceLock *version_map.WorkspaceLock) *InstallData {
	return &InstallData{
		Lock:              workspaceLock,
		allAvailable:      make(version_map.VersionListMap),
		RecentlyInstalled: make(version_map.ResolvedVersionListMap),
		AlreadyInstalled:  make(version_map.ResolvedVersionListMap),
	}
}

// GetAvailableUpdates returns a map of all installed mods which are not in the lock file
func (s *InstallData) GetAvailableUpdates() (version_map.DependencyVersionMap, error) {
	res := make(version_map.DependencyVersionMap)
	for parent, deps := range s.Lock.InstallCache {
		for name, constraints := range deps {
			availableVersions, err := s.getAvailableModVersions(name)
			if err != nil {
				return nil, err
			}
			for _, resolvedConstraint := range constraints {
				constraint, _ := version.NewConstraint(resolvedConstraint.Constraint)
				var latestVersion = getVersionSatisfyingConstraint(constraint, availableVersions)
				if latestVersion.GreaterThan(resolvedConstraint.Version) {
					res.Add(name, latestVersion, constraint, parent)
				}
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
	s.Lock.InstallCache.Add(dependency.Name, dependency.Version, modVersion.Constraint, parent.Name())
	// update list items installed by this installer
	s.RecentlyInstalled.Add(dependency.Name, &version_map.ResolvedVersionConstraint{
		Name:       dependency.Name,
		Version:    dependency.Version,
		Constraint: dependency.Constraint.Original,
	})
}

// addExisting is called when a dependency is satisfied by a mod which is already installed
func (s *InstallData) addExisting(name string, version *semver.Version, parent *modconfig.Mod) {
	// update lock
	modVersion := parent.Requires.GetModDependency(name)
	// update list of already installed items
	s.AlreadyInstalled.Add(name, &version_map.ResolvedVersionConstraint{
		Name:       name,
		Version:    version,
		Constraint: modVersion.Constraint.Original,
	})
}

//
// retrieve all available mod versions from our cache, or from Git if not yet cached
func (s *InstallData) getAvailableModVersions(modName string) ([]*semver.Version, error) {
	// have we already loaded the versions for this mod
	availableVersions, ok := s.allAvailable[modName]
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
	s.allAvailable[modName] = availableVersions

	return availableVersions, nil
}
