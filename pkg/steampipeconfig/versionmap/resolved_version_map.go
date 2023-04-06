package versionmap

// ResolvedVersionMap represents a map of ResolvedVersionConstraint, keyed by dependency name
type ResolvedVersionMap map[string]*ResolvedVersionConstraint

func (m ResolvedVersionMap) Add(name string, constraint *ResolvedVersionConstraint) {
	m[name] = constraint
}

func (m ResolvedVersionMap) Remove(name string) {
	delete(m, name)
}

// ToVersionListMap converts this map into a ResolvedVersionListMap
func (m ResolvedVersionMap) ToVersionListMap() ResolvedVersionListMap {
	res := make(ResolvedVersionListMap, len(m))
	for k, v := range m {
		res.Add(k, v)
	}
	return res
}

func (m ResolvedVersionMap) ToDependencyPathMap() map[string]struct{} {
	res := make(map[string]struct{}, len(m))
	for _, c := range m {
		res[c.DependencyPath()] = struct{}{}
	}
	return res
}
