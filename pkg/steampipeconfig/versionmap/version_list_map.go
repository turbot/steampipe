package versionmap

import (
	"sort"

	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// VersionListMap is a map keyed by dependency name storing a list of versions for each dependency
type VersionListMap map[string]semver.Collection

func (i VersionListMap) Add(name string, version *semver.Version) {
	versions := append(i[name], version)
	// reverse sort the versions
	sort.Sort(sort.Reverse(versions))
	i[name] = versions

}

// FlatMap converts the VersionListMap map into a bool map keyed by qualified dependency name
func (m VersionListMap) FlatMap() map[string]bool {
	var res = make(map[string]bool)
	for name, versions := range m {
		for _, version := range versions {
			key := modconfig.ModVersionFullName(name, version)
			res[key] = true
		}
	}
	return res
}
