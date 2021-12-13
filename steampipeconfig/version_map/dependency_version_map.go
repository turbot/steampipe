package version_map

import (
	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type DependencyVersionMap map[string]ResolvedVersionMap

// Add adds a dependency to the list of items installed for the given parent
func (m DependencyVersionMap) Add(dependencyName string, dependencyVersion *semver.Version, constraintString string, parentName string) {
	// get the map for this parent
	parentItems := m[parentName]
	// create if needed
	if parentItems == nil {
		parentItems = make(ResolvedVersionMap)
	}
	// add the dependency
	parentItems.Add(dependencyName, &ResolvedVersionConstraint{dependencyName, dependencyVersion, constraintString})
	// save
	m[parentName] = parentItems
}

// FlatMap converts the DependencyVersionMap into a ResolvedVersionMap, keyed by full name
func (m DependencyVersionMap) FlatMap() ResolvedVersionMap {
	res := make(ResolvedVersionMap)
	for _, deps := range m {
		for _, dep := range deps {
			res[modconfig.ModVersionFullName(dep.Name, dep.Version)] = dep
		}
	}
	return res
}
