package modconfig

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
)

type ResolvedVersionConstraint struct {
	Version    *semver.Version
	Constraint string
}

// VersionMap represents a map of semver versions, keyed by dependency name
type VersionMap map[string]*semver.Version

// ResolvedVersionMap represents a map of ResolvedVersionConstraint, keyed by dependency name
type ResolvedVersionMap map[string]*ResolvedVersionConstraint

// VersionsMap is a map keyed by dependency name storing a list of versions for each dependency
type VersionsMap map[string]semver.Collection

func (i VersionsMap) GetVersionSatisfyingRequirement(requiredVersion *ModVersionConstraint) *semver.Version {
	// is this dependency installed
	versions, ok := i[requiredVersion.Name]
	if !ok {
		return nil
	}
	for _, v := range versions {
		if requiredVersion.Constraint.Check(v) {
			return v
		}
	}
	return nil
}

func (i VersionsMap) Add(name string, version *semver.Version) {
	versions := append(i[name], version)
	// reverse sort the versions
	sort.Sort(sort.Reverse(versions))
	i[name] = versions

}

// FlatMap converts the VersionsMap map into a bool map keyed by qualified dependency name
func (m VersionsMap) FlatMap() map[string]bool {
	var res = make(map[string]bool)
	for name, versions := range m {
		for _, version := range versions {
			key := fmt.Sprintf("%s@%s", name, version)
			res[key] = true
		}
	}
	return res
}
