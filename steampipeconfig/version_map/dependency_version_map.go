package version_map

import (
	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/version"
)

type DependencyVersionMap map[string]ResolvedVersionMap

// Add adds a dependency to the list of items installed for the given parent
func (m DependencyVersionMap) Add(dependencyName string, dependencyVersion *semver.Version, constraint *version.Constraints, parentName string) {
	// get the map for this parent
	parentItems := m[parentName]
	// create if needed
	if parentItems == nil {
		parentItems = make(ResolvedVersionMap)
	}
	// add the dependency
	parentItems.Add(dependencyName, &ResolvedVersionConstraint{dependencyName, dependencyVersion, constraint.Original})
	// save
	m[parentName] = parentItems
}
