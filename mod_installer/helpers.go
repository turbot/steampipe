package mod_installer

import (
	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/version_helpers"
)

func getVersionSatisfyingConstraint(constraint *version_helpers.Constraints, availableVersions []*semver.Version) *semver.Version {
	// search the reverse sorted versions, finding the highest version which satisfies ALL constraints
	for _, version := range availableVersions {
		if constraint.Check(version) {
			return version
		}
	}
	return nil
}
