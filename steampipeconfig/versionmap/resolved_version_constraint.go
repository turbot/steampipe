package versionmap

import (
	"github.com/Masterminds/semver"
)

type ResolvedVersionConstraint struct {
	Name       string
	Alias      string
	Version    *semver.Version
	Constraint string
}

func NewResolvedVersionConstraint(name, alias string, version *semver.Version, constraintString string) *ResolvedVersionConstraint {
	return &ResolvedVersionConstraint{Name: name, Alias: alias, Version: version, Constraint: constraintString}
}

func (c ResolvedVersionConstraint) Equals(other *ResolvedVersionConstraint) bool {
	return c.Name == other.Name &&
		c.Version.Equal(other.Version) &&
		c.Constraint == other.Constraint
}

func (c ResolvedVersionConstraint) IsPrerelease() bool {
	return c.Version.Prerelease() != "" || c.Version.Metadata() != ""
}
