package versionmap

import (
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/xlab/treeprint"
	"golang.org/x/exp/maps"
)

type DependencyVersionMap map[string]ResolvedVersionMap

// Add adds a dependency to the list of items installed for the given parent
func (m DependencyVersionMap) Add(dependencyName, alias string, dependencyVersion *semver.Version, constraintString string, parentName string) {
	// get the map for this parent
	parentItems := m[parentName]
	// create if needed
	if parentItems == nil {
		parentItems = make(ResolvedVersionMap)
	}
	// add the dependency
	parentItems.Add(dependencyName, NewResolvedVersionConstraint(dependencyName, alias, dependencyVersion, constraintString))
	// save
	m[parentName] = parentItems
}

// FlatMap converts the DependencyVersionMap into a ResolvedVersionMap, keyed by mod dependency path
func (m DependencyVersionMap) FlatMap() ResolvedVersionMap {
	res := make(ResolvedVersionMap)
	for _, deps := range m {
		for _, dep := range deps {
			res[modconfig.BuildModDependencyPath(dep.Name, dep.Version)] = dep
		}
	}
	return res
}

func (m DependencyVersionMap) GetDependencyTree(rootName string) treeprint.Tree {
	tree := treeprint.NewWithRoot(rootName)
	m.buildTree(rootName, tree)
	return tree
}

func (m DependencyVersionMap) buildTree(name string, tree treeprint.Tree) {
	deps := m[name]
	depNames := maps.Keys(deps)
	sort.Strings(depNames)
	for _, name := range depNames {
		version := deps[name]
		fullName := modconfig.BuildModDependencyPath(name, version.Version)
		child := tree.AddBranch(fullName)
		// if there are children add them
		m.buildTree(fullName, child)
	}
}

// GetMissingFromOther returns a map of dependencies which exit in this map but not 'other'
func (m DependencyVersionMap) GetMissingFromOther(other DependencyVersionMap) DependencyVersionMap {
	res := make(DependencyVersionMap)
	for parent, deps := range m {
		otherDeps := other[parent]
		if otherDeps == nil {
			otherDeps = make(ResolvedVersionMap)
		}
		for name, dep := range deps {
			if _, ok := otherDeps[name]; !ok {
				res.Add(dep.Name, dep.Alias, dep.Version, dep.Constraint, parent)
			}
		}
	}
	return res
}

func (m DependencyVersionMap) GetUpgradedInOther(other DependencyVersionMap) DependencyVersionMap {
	res := make(DependencyVersionMap)
	for parent, deps := range m {
		otherDeps := other[parent]
		if otherDeps == nil {
			otherDeps = make(ResolvedVersionMap)
		}
		for name, dep := range deps {
			if otherDep, ok := otherDeps[name]; ok {
				if otherDep.Version.GreaterThan(dep.Version) {
					res.Add(otherDep.Name, dep.Alias, otherDep.Version, otherDep.Constraint, parent)
				}
			}
		}
	}
	return res
}

func (m DependencyVersionMap) GetDowngradedInOther(other DependencyVersionMap) DependencyVersionMap {
	res := make(DependencyVersionMap)
	for parent, deps := range m {
		otherDeps := other[parent]
		if otherDeps == nil {
			otherDeps = make(ResolvedVersionMap)
		}
		for name, dep := range deps {
			if otherDep, ok := otherDeps[name]; ok {
				if otherDep.Version.LessThan(dep.Version) {
					res.Add(otherDep.Name, dep.Alias, otherDep.Version, otherDep.Constraint, parent)
				}
			}
		}
	}
	return res
}
