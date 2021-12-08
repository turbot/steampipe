package modconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/turbot/steampipe/version"

	"github.com/Masterminds/semver"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
)

// WorkspaceLock is a map of ModVersionMaps items keyed by the parent mod whose dependencies are installed
type WorkspaceLock map[string]ResolvedVersionMap

func LoadWorkspaceLock(workspacePath string) (WorkspaceLock, error) {
	lockPath := constants.WorkspaceLockPath(workspacePath)
	if !helpers.FileExists(lockPath) {
		return nil, nil
	}

	fileContent, err := os.ReadFile(lockPath)
	if err != nil {
		log.Printf("[TRACE] error reading %s: %s\n", lockPath, err.Error())
		return nil, err
	}
	var res = make(WorkspaceLock)
	err = json.Unmarshal(fileContent, &res)
	if err != nil {
		log.Printf("[TRACE] failed to unmarshal %s: %s\n", lockPath, err.Error())
		return nil, nil
	}

	return res, nil
}

// Add adds a dependency to the list of items installed for the given parent
func (l WorkspaceLock) Add(dependencyName string, dependencyVersion *semver.Version, constraint *version.Constraints, parentName string) {
	// get the map for this parent
	parentItems := l[parentName]
	// create if needed
	if parentItems == nil {
		parentItems = make(ResolvedVersionMap)
	}
	// add the dependency

	parentItems[dependencyName] = &ResolvedVersionConstraint{dependencyVersion, constraint.Original}
	// save
	l[parentName] = parentItems
}

func (l WorkspaceLock) Save(workspacePath string) error {
	content, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(constants.WorkspaceLockPath(workspacePath), content, 0644)
}

func (l WorkspaceLock) Delete(workspacePath string) error {
	return os.Remove(constants.WorkspaceLockPath(workspacePath))
}

func (l WorkspaceLock) GetLockedModVersion(requiredModVersion *ModVersionConstraint, parent *Mod) (*ModVersionConstraint, error) {
	parentDependencies := l[parent.Name()]
	if parentDependencies == nil {
		return nil, nil
	}
	// look for this mod in the lock file entries for this parent
	lockedVersion := parentDependencies[requiredModVersion.Name]
	if lockedVersion == nil {
		return nil, nil
	}
	// verify the locked version satisfies the version constraint
	if !requiredModVersion.Constraint.Check(lockedVersion.Version) {
		return nil, fmt.Errorf("failed to install dependencies for %s - locked version %s@%s does not meet the constraint %s", parent.Name(), ModVersionFullName(requiredModVersion.Name, lockedVersion.Version), requiredModVersion.Constraint.Original)
	}
	// create a new requiredModVersion using the locked version
	return NewModVersionConstraint(fmt.Sprintf("%s@%s", requiredModVersion.Name, lockedVersion))
}

// ContainsMod returns whether the lockfile contains any version of the given mod
func (l WorkspaceLock) ContainsMod(modName string) bool {
	for _, modVersionMap := range l {
		for name := range modVersionMap {
			if name == modName {
				return true
			}
		}
	}
	return false
}

// ContainsModVersion returns whether the lockfile contains the given mod version
func (l WorkspaceLock) ContainsModVersion(modName string, modVersion *semver.Version) bool {
	for _, modVersionMap := range l {
		for lockName, lockVersion := range modVersionMap {
			if lockName == modName && lockVersion.Version.Equal(modVersion) {
				return true
			}
		}
	}
	return false
}
