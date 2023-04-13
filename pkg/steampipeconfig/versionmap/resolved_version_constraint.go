package versionmap

import (
	"github.com/Masterminds/semver"
)

type ResolvedVersionConstraint struct {
	Name          string          `json:"name,omitempty"`
	Alias         string          `json:"alias,omitempty"`
	Version       *semver.Version `json:"version,omitempty"`
	Constraint    string          `json:"constraint,omitempty"`
	StructVersion int             `json:"struct_version,omitempty"`
}

func NewResolvedVersionConstraint(name, alias string, version *semver.Version, constraintString string) *ResolvedVersionConstraint {
	return &ResolvedVersionConstraint{
		Name:          name,
		Alias:         alias,
		Version:       version,
		Constraint:    constraintString,
		StructVersion: WorkspaceLockStructVersion,
	}
}

func (c ResolvedVersionConstraint) Equals(other *ResolvedVersionConstraint) bool {
	return c.Name == other.Name &&
		c.Version.Equal(other.Version) &&
		c.Constraint == other.Constraint
}

func (c ResolvedVersionConstraint) IsPrerelease() bool {
	return c.Version.Prerelease() != "" || c.Version.Metadata() != ""
}
