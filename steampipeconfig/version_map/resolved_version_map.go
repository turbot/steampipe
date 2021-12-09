package version_map

// ResolvedVersionMap represents a map of ResolvedVersionConstraint, keyed by dependency name
type ResolvedVersionMap map[string]*ResolvedVersionConstraint
