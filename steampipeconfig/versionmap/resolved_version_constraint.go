package versionmap

import (
	"strings"

	"github.com/Masterminds/semver"
)

type ResolvedVersionConstraint struct {
	Name       string
	ShortName  string
	Version    *semver.Version
	Constraint string
}

func NewResolvedVersionConstraint(name string, version *semver.Version, constraintString string) *ResolvedVersionConstraint {
	shortName := getModShortName(name)
	return &ResolvedVersionConstraint{Name: name, ShortName: shortName, Version: version, Constraint: constraintString}
}

func getModShortName(name string) string {
	split := strings.Split(name, "/")
	return split[len(split)-1]
}

func (c ResolvedVersionConstraint) Equals(other *ResolvedVersionConstraint) bool {
	return c.Name == other.Name &&
		c.Version.Equal(other.Version) &&
		c.Constraint == other.Constraint
}

func (c ResolvedVersionConstraint) IsPrerelease() bool {
	return c.Version.Prerelease() != "" || c.Version.Metadata() != ""
}
