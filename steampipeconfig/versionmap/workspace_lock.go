package versionmap

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/turbot/steampipe/filepaths"

	"github.com/Masterminds/semver"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/versionhelpers"
)

// WorkspaceLock is a map of ModVersionMaps items keyed by the parent mod whose dependencies are installed
type WorkspaceLock struct {
	WorkspacePath       string
	ModInstallationPath string

	installCache    DependencyVersionMap
	missingVersions DependencyVersionMap
	installedMods   VersionListMap
	mutex           sync.Mutex
}

// EmptyWorkspaceLock creates a new empty workspace lock based,
// sharing workspace path and installedMods with 'existingLock'
func EmptyWorkspaceLock(existingLock *WorkspaceLock) *WorkspaceLock {
	return &WorkspaceLock{
		WorkspacePath:       existingLock.WorkspacePath,
		ModInstallationPath: filepaths.WorkspaceModPath(existingLock.WorkspacePath),
		installCache:        make(DependencyVersionMap),
		missingVersions:     make(DependencyVersionMap),
		installedMods:       existingLock.installedMods,
	}
}

func LoadWorkspaceLock(workspacePath string) (*WorkspaceLock, error) {
	var installCache = make(DependencyVersionMap)
	lockPath := filepaths.WorkspaceLockPath(workspacePath)
	if helpers.FileExists(lockPath) {

		fileContent, err := os.ReadFile(lockPath)
		if err != nil {
			log.Printf("[TRACE] error reading %s: %s\n", lockPath, err.Error())
			return nil, err
		}
		err = json.Unmarshal(fileContent, &installCache)
		if err != nil {
			log.Printf("[TRACE] failed to unmarshal %s: %s\n", lockPath, err.Error())
			return nil, nil
		}
	}
	res := &WorkspaceLock{
		WorkspacePath:       workspacePath,
		ModInstallationPath: filepaths.WorkspaceModPath(workspacePath),
		installCache:        installCache,
		missingVersions:     make(DependencyVersionMap),
	}

	if err := res.populateInstalledMods(); err != nil {
		return nil, err
	}

	// populate the MissingVersions
	// (this removes missing items from the install cache)
	res.setMissing()

	return res, nil
}

// GetUnreferencedMods returns a map of all installed mods which are not in the lock file
func (l *WorkspaceLock) GetUnreferencedMods() VersionListMap {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var unreferencedVersions = make(VersionListMap)
	for name, versions := range l.installedMods {
		for _, version := range versions {
			if !l.ContainsModVersion(name, version) {
				unreferencedVersions.Add(name, version)
			}
		}
	}
	return unreferencedVersions
}

func (l *WorkspaceLock) Save() error {
	if len(l.installCache) == 0 {
		// ignore error
		l.Delete()
		return nil
	}
	content, err := json.MarshalIndent(l.installCache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepaths.WorkspaceLockPath(l.WorkspacePath), content, 0644)
}

// Delete deletes the lock file
func (l *WorkspaceLock) Delete() error {
	if helpers.FileExists(filepaths.WorkspaceLockPath(l.WorkspacePath)) {
		return os.Remove(filepaths.WorkspaceLockPath(l.WorkspacePath))
	}
	return nil
}

// Add adds the dependency to our install cache, in a thread safe way
func (l *WorkspaceLock) Add(dependencyName string, dependencyVersion *semver.Version, constraintString string, parentName string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.installCache.Add(dependencyName, dependencyVersion, constraintString, parentName)
}

func (l *WorkspaceLock) GetDependenciesForMod(modName string) ResolvedVersionMap {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.installCache[modName]
}

// GetMod looks for a lock file entry matching the given mod name
func (l *WorkspaceLock) GetMod(modName string, parent *modconfig.Mod) *ResolvedVersionConstraint {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if parentDependencies := l.installCache[parent.GetModDependencyPath()]; parentDependencies != nil {
		// look for this mod in the lock file entries for this parent
		return parentDependencies[modName]
	}
	return nil
}

// GetLockedModVersions builds a ResolvedVersionListMap with the resolved versions
// for each item of the given VersionConstraintMap found in the lock file
func (l *WorkspaceLock) GetLockedModVersions(mods VersionConstraintMap, parent *modconfig.Mod) (ResolvedVersionListMap, error) {
	// do not lock as we call back into a locked function

	var res = make(ResolvedVersionListMap)
	for name, constraint := range mods {
		resolvedConstraint, err := l.GetLockedModVersion(constraint, parent)
		if err != nil {
			return nil, err
		}
		if resolvedConstraint != nil {
			res.Add(name, resolvedConstraint)
		}
	}
	return res, nil
}

// GetLockedModVersion looks for a lock file entry matching the required constraint and returns nil if not found
func (l *WorkspaceLock) GetLockedModVersion(requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod) (*ResolvedVersionConstraint, error) {
	// do not lock as we call back into a locked function

	lockedVersion := l.GetMod(requiredModVersion.Name, parent)
	if lockedVersion == nil {
		return nil, nil
	}

	// verify the locked version satisfies the version constraint
	if !requiredModVersion.Constraint.Check(lockedVersion.Version) {
		return nil, nil
	}

	return lockedVersion, nil
}

// EnsureLockedModVersion looks for a lock file entry matching the required mod name,
// and verify it satisfies the version constraint
// NOTE: throws an error if the version constraint is not satisfied
func (l *WorkspaceLock) EnsureLockedModVersion(requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod) (*ResolvedVersionConstraint, error) {
	// do not lock as we call back into a locked function

	lockedVersion := l.GetMod(requiredModVersion.Name, parent)
	if lockedVersion == nil {
		return nil, nil
	}

	// verify the locked version satisfies the version constraint
	if !requiredModVersion.Constraint.Check(lockedVersion.Version) {
		return nil, fmt.Errorf("failed to resolvedependencies for %s - locked version %s does not meet the constraint %s", parent.GetModDependencyPath(), modconfig.ModVersionFullName(requiredModVersion.Name, lockedVersion.Version), requiredModVersion.Constraint.Original)
	}

	return lockedVersion, nil
}

// GetLockedModVersionConstraint looks for a lock file entry matching the required mod version and if found,
// returns it in the form of a ModVersionConstraint
func (l *WorkspaceLock) GetLockedModVersionConstraint(requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod) (*modconfig.ModVersionConstraint, error) {
	// do not lock as we call back into a locked function

	lockedVersion, err := l.EnsureLockedModVersion(requiredModVersion, parent)
	if err != nil {
		// EnsureLockedModVersion returns an error if the locked version does not satisfy the requirement
		return nil, err
	}
	if lockedVersion == nil {
		// EnsureLockedModVersion returns nil if no locked version is found
		return nil, nil
	}
	// create a new ModVersionConstraint using the locked version
	lockedVersionFullName := modconfig.ModVersionFullName(requiredModVersion.Name, lockedVersion.Version)
	return modconfig.NewModVersionConstraint(lockedVersionFullName)
}

// ContainsModVersion returns whether the lockfile contains the given mod version
func (l *WorkspaceLock) ContainsModVersion(modName string, modVersion *semver.Version) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, modVersionMap := range l.installCache {
		for lockName, lockVersion := range modVersionMap {
			// TODO consider handling of metadata
			if lockName == modName && lockVersion.Version.Equal(modVersion) && lockVersion.Version.Metadata() == modVersion.Metadata() {
				return true
			}
		}
	}
	return false
}

func (l *WorkspaceLock) ContainsModConstraint(modName string, constraint *versionhelpers.Constraints) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, modVersionMap := range l.installCache {
		for lockName, lockVersion := range modVersionMap {
			if lockName == modName && lockVersion.Constraint == constraint.Original {
				return true
			}
		}
	}
	return false
}

func (l *WorkspaceLock) GetMissingFromOther(other *WorkspaceLock) DependencyVersionMap {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	res := make(DependencyVersionMap)
	for parent, deps := range l.installCache {
		otherDeps := other.GetDependenciesForMod(parent)
		if otherDeps == nil {
			otherDeps = make(ResolvedVersionMap)
		}
		for name, dep := range deps {
			if _, ok := otherDeps[name]; !ok {
				res.Add(dep.Name, dep.Version, dep.Constraint, parent)
			}
		}
	}
	return res
}

func (l *WorkspaceLock) GetUpgradedInOther(other *WorkspaceLock) DependencyVersionMap {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	res := make(DependencyVersionMap)
	for parent, deps := range l.installCache {
		otherDeps := other.GetDependenciesForMod(parent)
		if otherDeps == nil {
			otherDeps = make(ResolvedVersionMap)
		}
		for name, dep := range deps {
			if otherDep, ok := otherDeps[name]; ok {
				if otherDep.Version.GreaterThan(dep.Version) {
					res.Add(otherDep.Name, otherDep.Version, otherDep.Constraint, parent)
				}
			}
		}
	}
	return res
}

func (l *WorkspaceLock) GetDowngradedInOther(other *WorkspaceLock) DependencyVersionMap {
	res := make(DependencyVersionMap)
	for parent, deps := range l.installCache {
		otherDeps := other.GetDependenciesForMod(parent)
		if otherDeps == nil {
			otherDeps = make(ResolvedVersionMap)
		}
		for name, dep := range deps {
			if otherDep, ok := otherDeps[name]; ok {
				if otherDep.Version.LessThan(dep.Version) {
					res.Add(otherDep.Name, otherDep.Version, otherDep.Constraint, parent)
				}
			}
		}
	}
	return res
}

// Incomplete returned whether there are any missing dependencies
// (i.e. they exist in the lock file but ate not installed)
func (l *WorkspaceLock) Incomplete() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return len(l.missingVersions) > 0
}

// Empty returns whether the install cache is empty
func (l *WorkspaceLock) Empty() bool {
	if l == nil {
		return true
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return len(l.installCache) == 0
}

// populateInstalledMods returns a map installed mods, and the versions installed for each
func (l *WorkspaceLock) populateInstalledMods() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	// recursively search for all the mod.sp files under the .steampipe/mods folder, then build the mod name from the file path
	modFiles, err := filehelpers.ListFiles(l.ModInstallationPath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesRecursive,
		Include: []string{"**/mod.sp"},
	})
	if err != nil {
		return err
	}

	// create result map - a list of version for each mod
	installedMods := make(VersionListMap, len(modFiles))
	// collect errors
	var errors []error

	for _, modfilePath := range modFiles {
		// try to parse the mon name and version form the parent folder of the modfile
		modName, version, err := l.parseModPath(modfilePath)
		if err != nil {
			// if we fail to parse, just ignore this modfile
			// - it's parent is not a valid mod installation folder so it is probably a child folder of a mod
			continue
		}
		// add this mod version to the map
		installedMods.Add(modName, version)
	}

	if len(errors) > 0 {
		return utils.CombineErrors(errors...)
	}
	l.installedMods = installedMods
	return nil
}

// identify mods which are in installCache but not installed
// move them from installCache into missingVersions
func (l *WorkspaceLock) setMissing() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// create a map of full modname to bool to allow simple checking
	flatInstalled := l.installedMods.FlatMap()

	for parent, deps := range l.installCache {
		// deps is a map of dep name to resolved contraint list
		// flatten and iterate

		for name, resolvedConstraint := range deps {
			fullName := modconfig.ModVersionFullName(name, resolvedConstraint.Version)

			if !flatInstalled[fullName] {
				// get the mod name from the constraint (fullName includes the version)
				name := resolvedConstraint.Name
				// remove this item from the install cache and add into missing
				l.missingVersions.Add(name, resolvedConstraint.Version, resolvedConstraint.Constraint, parent)
				l.installCache[parent].Remove(name)
			}
		}
	}
}

// extract the mod name and version from the modfile path
func (l *WorkspaceLock) parseModPath(modfilePath string) (modName string, modVersion *semver.Version, err error) {
	modFullName, err := filepath.Rel(l.ModInstallationPath, filepath.Dir(modfilePath))
	if err != nil {
		return
	}
	return modconfig.ParseModFullName(modFullName)
}
