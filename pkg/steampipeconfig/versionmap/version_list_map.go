package versionmap

import (
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/turbot/pipe-fittings/modconfig"
)

// VersionListMap is a map keyed by dependency name storing a list of versions for each dependency
type VersionListMap map[string]semver.Collection

func (m VersionListMap) Add(name string, version *semver.Version) {
	versions := append(m[name], version)
	// reverse sort the versions
	sort.Sort(sort.Reverse(versions))
	m[name] = versions

}

// FlatMap converts the VersionListMap map into a bool map keyed by qualified dependency name
func (m VersionListMap) FlatMap() map[string]bool {
	var res = make(map[string]bool)
	for name, versions := range m {
		for _, version := range versions {
			key := modconfig.BuildModDependencyPath(name, version)
			res[key] = true
		}
	}
	return res
}
