package version_map

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/version"

	"github.com/Masterminds/semver"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
)

// WorkspaceLock is a map of ModVersionMaps items keyed by the parent mod whose dependencies are installed
type WorkspaceLock struct {
	WorkspacePath        string
	InstallCache         DependencyVersionMap
	MissingVersions      DependencyVersionMap
	UnreferencedVersions VersionListMap
	modsPath             string
}

func LoadWorkspaceLock(workspacePath string) (*WorkspaceLock, error) {
	lockPath := constants.WorkspaceLockPath(workspacePath)
	if !helpers.FileExists(lockPath) {
		return nil, nil
	}

	fileContent, err := os.ReadFile(lockPath)
	if err != nil {
		log.Printf("[TRACE] error reading %s: %s\n", lockPath, err.Error())
		return nil, err
	}
	var installCache = make(DependencyVersionMap)
	err = json.Unmarshal(fileContent, &installCache)
	if err != nil {
		log.Printf("[TRACE] failed to unmarshal %s: %s\n", lockPath, err.Error())
		return nil, nil
	}
	res := &WorkspaceLock{
		WorkspacePath:        workspacePath,
		modsPath:             constants.WorkspaceModPath(workspacePath),
		InstallCache:         installCache,
		MissingVersions:      make(DependencyVersionMap),
		UnreferencedVersions: make(VersionListMap),
	}

	if err := res.validate(); err != nil {
		return nil, err
	}
	return res, nil
}

// populate MissingVersions and UnreferencedVersions
func (l *WorkspaceLock) validate() error {
	installedMods, err := l.getInstalledMods()
	if err != nil {
		return err
	}
	l.setUnreferenced(installedMods)
	l.setMissing(installedMods)
	return nil
}

// getInstalledMods returns a map installed mods, and the versions installed for each
func (l *WorkspaceLock) getInstalledMods() (VersionListMap, error) {
	// recursively search for all the mod.sp files under the .steampipe/mods folder, then build the mod name from the file path
	modFiles, err := filehelpers.ListFiles(l.modsPath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesRecursive,
		Include: []string{"**/mod.sp"},
	})
	if err != nil {
		return nil, err
	}

	// create result map - a list of version for each mod
	installedMods := make(VersionListMap, len(modFiles))
	// collect errors
	var errors []error

	for _, modfilePath := range modFiles {
		modName, version, err := l.parseModPath(modfilePath)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		// add this mod version to the map
		installedMods.Add(modName, version)
	}

	if len(errors) > 0 {
		return nil, utils.CombineErrors(errors...)
	}
	return installedMods, nil
}

// identify map of all installed mods which are not in the lock file
func (l *WorkspaceLock) setUnreferenced(installedMods VersionListMap) {
	for name, versions := range installedMods {
		for _, version := range versions {
			if !l.ContainsModVersion(name, version) {
				l.UnreferencedVersions.Add(name, version)
			}
		}
	}
}

// identify mods which are in tInstallCache but not installed
// move them from InstallCache into MissingVersions
func (l *WorkspaceLock) setMissing(installedMods VersionListMap) {
	// create a map of full modname to bool to allow simple checking
	flatInstalled := installedMods.FlatMap()

	for parent, deps := range l.InstallCache {
		// deps is a map of dep name to resolved contraint list
		// flatten and iterate
		for name, resolvedConstraint := range deps.FlatMap() {
			// build full name and check map of installed
			fullName := modconfig.ModVersionFullName(name, resolvedConstraint.Version)
			if !flatInstalled[fullName] {
				// remove this item from th einstall cache and add into missing
				l.MissingVersions[parent].Add(name, resolvedConstraint)
				l.InstallCache[parent].Remove(name, resolvedConstraint)
			}
		}
	}
}

// extract the mod name and version from the modfile path
func (l *WorkspaceLock) parseModPath(modfilePath string) (modName string, modVersion *semver.Version, err error) {
	modFullName, err := filepath.Rel(l.modsPath, filepath.Dir(modfilePath))
	if err != nil {
		return
	}
	return modconfig.ParseModFullName(modFullName)
}

func (l *WorkspaceLock) Save(workspacePath string) error {
	content, err := json.MarshalIndent(l.InstallCache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(constants.WorkspaceLockPath(workspacePath), content, 0644)
}

func (l *WorkspaceLock) Delete(workspacePath string) error {
	return os.Remove(constants.WorkspaceLockPath(workspacePath))
}

func (l *WorkspaceLock) GetLockedModVersion(requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod) (*modconfig.ModVersionConstraint, error) {
	parentDependencies := l.InstallCache[parent.Name()]
	if parentDependencies == nil {
		return nil, nil
	}
	// look for this mod in the lock file entries for this parent
	lockedVersions := parentDependencies[requiredModVersion.Name]
	if len(lockedVersions) == 0 {
		return nil, nil
	}
	// NOTE: for now we only support a single version of each mod per parent
	// when we support aliases this restriction can be removed
	if len(lockedVersions) > 1 {
		return nil, fmt.Errorf("parent %s has more than 1 version of dependency %s", parent.Name(), requiredModVersion.Name)
	}
	lockedVersion := lockedVersions[0]
	// verify the locked version satisfies the version constraint
	if !requiredModVersion.Constraint.Check(lockedVersion.Version) {
		return nil, fmt.Errorf("failed to install dependencies for %s - locked version %s@%s does not meet the constraint %s", parent.Name(), modconfig.ModVersionFullName(requiredModVersion.Name, lockedVersion.Version), requiredModVersion.Constraint.Original)
	}
	// create a new requiredModVersion using the locked version
	return modconfig.NewModVersionConstraint(fmt.Sprintf("%s@%s", requiredModVersion.Name, lockedVersion))
}

// ContainsMod returns whether the lockfile contains any version of the given mod
func (l *WorkspaceLock) ContainsMod(modName string) bool {
	for _, modVersionMap := range l.InstallCache {
		for name := range modVersionMap {
			if name == modName {
				return true
			}
		}
	}
	return false
}

// ContainsModVersion returns whether the lockfile contains the given mod version
func (l *WorkspaceLock) ContainsModVersion(modName string, modVersion *semver.Version) bool {
	for _, modVersionMap := range l.InstallCache {
		for lockName, lockVersions := range modVersionMap {
			// we only support a single version at the moment but iterate anyway - we validate elsewhere
			for _, lockVersion := range lockVersions {
				if lockName == modName && lockVersion.Version.Equal(modVersion) {
					return true
				}
			}
		}
	}
	return false
}

func (l *WorkspaceLock) ContainsModConstraint(modName string, constraint *version.Constraints) bool {
	for _, modVersionMap := range l.InstallCache {
		for lockName, lockVersions := range modVersionMap {
			// we only support a single version at the moment but iterate anyway - we validate elsewhere
			for _, lockVersion := range lockVersions {
				if lockName == modName && lockVersion.Constraint == constraint.Original {
					return true
				}
			}
		}
	}
	return false
}
