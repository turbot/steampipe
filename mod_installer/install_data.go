package mod_installer

import (
	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type InstallData struct {
	// record of the full dependency tree
	Lock modconfig.WorkspaceLock

	// all installed mod versions (some of which may not be required)
	AllInstalled InstalledModMap
	// mod versions which are installed but not needed
	Unreferenced InstalledModMap
	// ALL the available versions for each dependency mod(we populate this in a lazy fashion)
	AllAvailable InstalledModMap

	// list of dependencies installed by recent install operation
	RecentlyInstalled []string
	// list of dependencies which were already installed
	AlreadyInstalled []string
}

func NewInstallData(installedMods InstalledModMap, workspaceLock modconfig.WorkspaceLock) *InstallData {
	return &InstallData{
		Lock:         workspaceLock,
		AllInstalled: installedMods,
		Unreferenced: make(InstalledModMap),
		AllAvailable: make(InstalledModMap),
	}
}

// onModInstalled is called when a dependency is satisfied by installing a mod version
func (s *InstallData) onModInstalled(dependency *ResolvedModRef, parent *modconfig.Mod) {
	// update lock
	s.Lock.Add(parent.Name(), dependency.Name, dependency.Version)
	// update list of all installed mods
	s.AllInstalled.Add(dependency.Name, dependency.Version)
	// update list items installed by this installer
	s.RecentlyInstalled = append(s.RecentlyInstalled, dependency.FullName())
}

// addExisting is called when a dependency is satisfied by a mod which is alreayd installed
func (s *InstallData) addExisting(name string, version *semver.Version, parent *modconfig.Mod) {
	// update lock
	s.Lock.Add(parent.Name(), name, version)
	// update list of already installed items
	s.AlreadyInstalled = append(s.AlreadyInstalled, modVersionFullName(name, version))
}

// return a map of all installed mods which are not in the lock file
func (s *InstallData) getUnusedMods() InstalledModMap {
	var unusedModPaths = make(InstalledModMap)
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
