package modconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
)

// ModVersionMap represents a map of installed dependencies, keyed by name with versoin as the map value
type ModVersionMap map[string]*semver.Version

// WorkspaceLock is a map of ModVersionMaps items keyed by the parent mod whose dependencies are installed
type WorkspaceLock map[string]ModVersionMap

// Add adds a dependency to the list of items installed for the given parent
func (m WorkspaceLock) Add(parent, dependency string, dependencyVersion *semver.Version) {
	// get the map for this parent
	parentItems := m[parent]
	// create if needed
	if parentItems == nil {
		parentItems = make(ModVersionMap)
	}
	// add the dependency
	parentItems[dependency] = dependencyVersion
	// save
	m[parent] = parentItems
}

func (m WorkspaceLock) Save(workspacePath string) error {
	content, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(constants.WorkspaceLockPath(workspacePath), content, 0644)
}

func (m WorkspaceLock) GetLockedModVersion(requiredModVersion *ModVersionConstraint, parent *Mod) (*ModVersionConstraint, error) {
	parentDependencies := m[parent.Name()]
	if parentDependencies == nil {
		return nil, nil
	}
	// look for this mod in the lock file entries for this parent
	lockedVersion := parentDependencies[requiredModVersion.Name]
	// TODO if there is no locked version - error? require --update???
	if lockedVersion == nil {
		return nil, nil
	}
	// verify the locked version satisfies the version constraint
	if !requiredModVersion.Constraint.Check(lockedVersion) {
		// TODO ignore if update is true???
		// WHAT TO DO
		return nil, fmt.Errorf("failed to install dependencies for %s - locked version %s@%s does not meet the constraint %s", parent.Name(), requiredModVersion.Name, lockedVersion.Original(), requiredModVersion.Constraint.Original)
	}
	// create a new requiredModVersion using the locked version
	return NewModVersionConstraint(fmt.Sprintf("%s@%s", requiredModVersion.Name, lockedVersion))

}

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
