package version_map

import "github.com/Masterminds/semver"

type ResolvedVersionConstraint struct {
	Name       string
	Version    *semver.Version
	Constraint string
}

func (c ResolvedVersionConstraint) Equals(other *ResolvedVersionConstraint) bool {
	return c.Name == other.Name &&
		c.Version.Equal(other.Version) &&
		c.Constraint == other.Constraint
}
