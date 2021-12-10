package version_map

// ResolvedVersionMap represents a map of ResolvedVersionConstraint, keyed by dependency name
type ResolvedVersionMap map[string]*ResolvedVersionConstraint

func (m ResolvedVersionMap) Add(name string, constraint *ResolvedVersionConstraint) {
	m[name] = constraint
}

func (m ResolvedVersionMap) Remove(name string) {
	delete(m, name)
}
