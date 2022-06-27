package versionmap

import (
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// ResolvedVersionListMap represents a map of ResolvedVersionConstraint arrays, keyed by dependency name
type ResolvedVersionListMap map[string][]*ResolvedVersionConstraint

// Add appends the version constraint to the list for the given name
func (m ResolvedVersionListMap) Add(name string, versionConstraint *ResolvedVersionConstraint) {
	// if there is already an entry for the same name, replace it
	// TODO handle alias
	m[name] = []*ResolvedVersionConstraint{versionConstraint}
}

// Remove removes the given version constraint from the list for the given name
func (m ResolvedVersionListMap) Remove(name string, constraint *ResolvedVersionConstraint) {
	var res []*ResolvedVersionConstraint
	for _, c := range m[name] {
		if !c.Equals(constraint) {
			res = append(res, c)
		}
	}
	m[name] = res
}

// FlatMap converts the ResolvedVersionListMap map into a map keyed by the FULL dependency name (i.e. including version(
func (m ResolvedVersionListMap) FlatMap() map[string]*ResolvedVersionConstraint {
	var res = make(map[string]*ResolvedVersionConstraint)
	for name, versions := range m {
		for _, version := range versions {
			key := modconfig.ModVersionFullName(name, version.Version)
			res[key] = version
		}
	}
	return res
}

// FlatNames converts the ResolvedVersionListMap map into a string array of full names
func (m ResolvedVersionListMap) FlatNames() []string {
	var res []string
	for name, versions := range m {
		for _, version := range versions {
			res = append(res, modconfig.ModVersionFullName(name, version.Version))
		}
	}
	return res
}
